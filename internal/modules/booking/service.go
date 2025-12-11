package booking

import (
	"errors"
	"fmt"
	"math/rand"

	//"strings"
	"time"

	"ezytix-be/internal/models"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/modules/payment"
)

type BookingService interface {
	CreateOrder(userID uint, req CreateOrderRequest) (*BookingResponse, error)
}

type bookingService struct {
	repo           BookingRepository
	flightService  flight.FlightService
	paymentService payment.PaymentService
	authService    auth.AuthService // <--- Inject Auth Service
}

// Update Constructor: Tambahkan authService
func NewBookingService(
	repo BookingRepository,
	flightService flight.FlightService,
	paymentService payment.PaymentService,
	authService auth.AuthService, // <--- Param Baru
) BookingService {
	return &bookingService{
		repo:           repo,
		flightService:  flightService,
		paymentService: paymentService,
		authService:    authService,
	}
}

// ==========================================
// CORE LOGIC: CREATE ORDER
// ==========================================
func (s *bookingService) CreateOrder(userID uint, req CreateOrderRequest) (*BookingResponse, error) {
	user, err := s.authService.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	// 1. Generate Order ID (The Glue)
	// Format: ORD-{Time}-{Random} -> ORD-20231025-X82A
	orderID := fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), generateRandomString(4))

	var bookingsToSave []models.Booking
	var grandTotal float64
	var bookingResponses []BookingDetailResponse

	// Tentukan Trip Type (OneWay atau RoundTrip)
	tripType := models.TripTypeOneWay
	if len(req.Items) > 1 {
		tripType = models.TripTypeRoundTrip
	}

	// 2. Loop Setiap Flight yang dipesan (Pergi & Pulang)
	for _, item := range req.Items {
		// A. Validasi Flight & Ambil Harga (Source of Truth)
		flightData, err := s.flightService.GetFlightByID(item.FlightID)
		if err != nil {
			return nil, errors.New("flight not found")
		}

		// Cari Kelas yang dipilih user di Database
		var selectedClass *models.FlightClass
		for _, fc := range flightData.FlightClasses {
			if fc.SeatClass == item.SeatClass {
				selectedClass = &fc
				break
			}
		}
		if selectedClass == nil {
			return nil, errors.New("seat class not available for this flight")
		}

		// B. Hitung Harga per Segment
		passengerCount := len(item.Passengers)
		flightTotalPrice := selectedClass.Price * float64(passengerCount)
		grandTotal += flightTotalPrice

		// C. Buat Struktur Booking (Header)
		bookingCode := generatePNR() // Generate PNR Unik (EZY-XXXXX)
		
		booking := models.Booking{
			OrderID:         orderID,
			UserID:          userID,
			FlightID:        item.FlightID,
			BookingCode:     bookingCode,
			TripType:        tripType,
			TotalPassengers: passengerCount,
			TotalPrice:      flightTotalPrice, // Harga snapshot per flight
			Status:          models.BookingStatusPending,
		}

		// D. Buat Struktur Details (Penumpang)
		var details []models.BookingDetail
		for _, pReq := range item.Passengers {
			// Parsing DOB untuk menentukan Tipe Penumpang
			dobTime, _ := time.Parse("2006-01-02", pReq.DOB)
			passengerType := calculatePassengerType(dobTime)

			// Generate Nomor Tiket Unik
			ticketNum := fmt.Sprintf("%s-%s", bookingCode, generateRandomString(3)) // Contoh: EZY-882X-A01

			detail := models.BookingDetail{
				PassengerName:  pReq.FullName,
				PassengerTitle: pReq.Title,
				PassengerDOB:   dobTime,
				Nationality:    pReq.Nationality,
				PassengerType:  passengerType,
				
				// Dokumen (Pointer)
				PassportNumber: stringToPointer(pReq.PassportNumber),
				IssuingCountry: stringToPointer(pReq.IssuingCountry),
				ValidUntil:     dateToPointer(pReq.ValidUntil),

				// Snapshot Transaksi
				TicketNumber:   ticketNum,
				SeatClass:      item.SeatClass,
				Price:          selectedClass.Price, // Harga satuan saat beli
			}
			details = append(details, detail)
		}
		booking.Details = details
		bookingsToSave = append(bookingsToSave, booking)

		// [FIXED] Ambil Nama Kota dari Relasi Airport (No More Hardcode)
		originCity := flightData.OriginAirport.CityName
		destCity := flightData.DestinationAirport.CityName

		bookingResponses = append(bookingResponses, BookingDetailResponse{
			BookingCode:     bookingCode,
			FlightCode:      flightData.FlightCode,
			Origin:          originCity, // Real Data
			Destination:     destCity,   // Real Data
			DepartureTime:   flightData.DepartureTime,
			TotalPassengers: passengerCount,
			TotalPrice:      flightTotalPrice,
		})
	}

	// 3. EXECUTE ATOMIC TRANSACTION (Repo)
	// Di sini stok berkurang & data tersimpan. Kalau gagal, rollback semua.
	if err := s.repo.CreateOrder(bookingsToSave); err != nil {
		return nil, err
	}

	// 4. CALL PAYMENT SERVICE (Xendit)
	// Kita minta payment gateway buatkan tagihan untuk OrderID ini
	paymentReq := payment.CreatePaymentRequest{
		OrderID:     orderID,
		Amount:      grandTotal,
		Description: fmt.Sprintf("Flight Booking %s (%s)", orderID, tripType),
		PayerEmail:  user.Email, // [FIXED] Real Email dari DB User
		PayerName:   user.FullName, // Optional: biar invoice xendit ada namanya
	}

	paymentResp, err := s.paymentService.CreatePayment(paymentReq)
	if err != nil {
		// PENTING: Jika payment gateway error, idealnya kita batalkan booking (Rollback)
		// Tapi untuk MVP, kita biarkan booking PENDING, nanti Cron Job yang bersihkan.
		// Atau return error agar user coba lagi.
		return nil, errors.New("booking created but payment initiation failed: " + err.Error())
	}

	// 5. Final Response
	return &BookingResponse{
		OrderID:         orderID,
		TotalAmount:     grandTotal,
		Status:          models.BookingStatusPending,
		TransactionTime: time.Now(),
		PaymentURL:      paymentResp.PaymentURL, // Redirect user ke sini!
		Bookings:        bookingResponses,
	}, nil
}

// ==========================================
// HELPER FUNCTIONS
// ==========================================

func generateRandomString(n int) string {
	const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func generatePNR() string {
	// PNR biasanya 6 karakter alfanumerik (misal: EZY88X)
	return fmt.Sprintf("EZY%s", generateRandomString(5))
}

func calculatePassengerType(dob time.Time) string {
	now := time.Now()
	age := now.Year() - dob.Year()
	// Koreksi jika belum ulang tahun tahun ini
	if now.YearDay() < dob.YearDay() {
		age--
	}

	if age >= 12 {
		return models.PassengerTypeAdult
	} else if age >= 2 {
		return models.PassengerTypeChild
	}
	return models.PassengerTypeInfant
}

func stringToPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func dateToPointer(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}
	return &t
}