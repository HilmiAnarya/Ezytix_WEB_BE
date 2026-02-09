package booking

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ezytix-be/internal/models"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/utils"
	pdfprinter "ezytix-be/internal/utils/pdf_printer"

	"github.com/shopspring/decimal"
)

type BookingService interface {
	CreateOrder(userID uint, req CreateOrderRequest) (*BookingResponse, error)
	ProcessExpiredBookings() error
	GetUserBookings(userID uint) ([]MyBookingResponse, error)
	DownloadInvoice(ctx context.Context, bookingCode string) ([]byte, error)
	DownloadEticket(ctx context.Context, bookingCode string) ([]byte, error)
}

type bookingService struct {
	repo          BookingRepository
	flightService flight.FlightService
	// [REMOVED] paymentService
	authService   auth.AuthService
}

func NewBookingService(
	repo BookingRepository,
	flightService flight.FlightService,
	// [REMOVED] paymentService
	authService auth.AuthService,
) BookingService {
	return &bookingService{
		repo:          repo,
		flightService: flightService,
		// [REMOVED] paymentService: paymentService,
		authService:   authService,
	}
}

func (s *bookingService) CreateOrder(userID uint, req CreateOrderRequest) (*BookingResponse, error) {
	_, err := s.authService.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	orderID := fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), generateRandomString(4))

	var bookingsToSave []models.Booking
	grandTotal := decimal.Zero
	var bookingResponses []BookingDetailResponse

	tripType := models.TripTypeOneWay
	if len(req.Items) > 1 {
		tripType = models.TripTypeRoundTrip
	}

	// [STRICT EXPIRY] Set Expiry Awal (misal 55 menit dari sekarang)
	// User harus initiate payment sebelum waktu ini.
	expiryDuration := 55 * time.Minute
	expiryAt := time.Now().Add(expiryDuration)

	for _, item := range req.Items {
		flightData, err := s.flightService.GetFlightByID(item.FlightID) // Sesuaikan: GetFlightById atau GetFlightByID
		if err != nil {
			return nil, errors.New("flight not found")
		}

		var selectedClass *models.FlightClass
		for _, fc := range flightData.FlightClasses {
			if strings.EqualFold(fc.SeatClass, item.SeatClass) {
				selectedClass = &fc
				break
			}
		}
		if selectedClass == nil {
			return nil, errors.New("seat class not available for this flight")
		}

		passengerCountInt := int64(len(item.Passengers))
		passengerCountDec := decimal.NewFromInt(passengerCountInt)
		flightTotalPrice := selectedClass.Price.Mul(passengerCountDec)
		grandTotal = grandTotal.Add(flightTotalPrice)

		bookingCode := generatePNR()
		
		booking := models.Booking{
			OrderID:         orderID,
			UserID:          userID,
			FlightID:        item.FlightID,
			BookingCode:     bookingCode,
			TripType:        tripType,
			TotalPassengers: len(item.Passengers),
			TotalPrice:      flightTotalPrice,
			Status:          models.BookingStatusPending,
			
			// [NEW] Simpan Expiry Time ke Database
			ExpiredAt:       &expiryAt, 
			CreatedAt:       time.Now(),
		}

		var details []models.BookingDetail
		for _, pReq := range item.Passengers {
			dobTime, _ := time.Parse("2006-01-02", pReq.DOB)
			passengerType := calculatePassengerType(dobTime)
			ticketNum := fmt.Sprintf("%s-%s", bookingCode, generateRandomString(3))

			detail := models.BookingDetail{
				PassengerName:  pReq.FullName,
				PassengerTitle: pReq.Title,
				PassengerDOB:   dobTime,
				Nationality:    pReq.Nationality,
				PassengerType:  passengerType,
				PassportNumber: stringToPointer(pReq.PassportNumber),
				IssuingCountry: stringToPointer(pReq.IssuingCountry),
				ValidUntil:     dateToPointer(pReq.ValidUntil),
				TicketNumber:   ticketNum,
				SeatClass:      item.SeatClass,
				Price:          selectedClass.Price,
			}
			details = append(details, detail)
		}
		booking.Details = details
		bookingsToSave = append(bookingsToSave, booking)

		bookingResponses = append(bookingResponses, BookingDetailResponse{
			BookingCode:     bookingCode,
			FlightCode:      flightData.FlightCode,
			Origin:          flightData.OriginAirport.CityName,
			Destination:     flightData.DestinationAirport.CityName,
			DepartureTime:   flightData.DepartureTime,
			TotalPassengers: len(item.Passengers),
			TotalPrice:      flightTotalPrice,
		})
	}

	if err := s.repo.CreateOrder(bookingsToSave); err != nil {
		return nil, err
	}

	// [REMOVED] Logic CreatePayment dihapus total.
	
	return &BookingResponse{
		OrderID:         orderID,
		TotalAmount:     grandTotal,
		Status:          models.BookingStatusPending,
		TransactionTime: time.Now(),
		// [FIXED] ExpiryDate diisi dari variabel yang sama dengan DB
		ExpiryTime:      &expiryAt,
		Bookings:        bookingResponses,
	}, nil
}

func (s *bookingService) ProcessExpiredBookings() error {
	log.Println("[CRON] --- Starting Scheduler Job ---")

	// TUGAS 1: Batalkan PENDING yang sudah lewat ExpiredAt
	// Logic baru: Cukup cek NOW() > expired_at
	
	expiredPendingBookings, err := s.repo.GetExpiredBookings(time.Now())
	if err != nil {
		log.Printf("[CRON] Error fetching pending expired bookings: %v\n", err)
	} else {
		if len(expiredPendingBookings) > 0 {
			log.Printf("[CRON] Found %d pending bookings to cancel.\n", len(expiredPendingBookings))
			for _, booking := range expiredPendingBookings {
				err := s.repo.CancelBookingAtomic(&booking)
				if err != nil {
					log.Printf("[CRON] Failed to cancel Pending Booking ID %d: %v\n", booking.ID, err)
					continue 
				}
				log.Printf("[CRON] Cancelled Pending Booking ID %d (Stock restored).\n", booking.ID)
			}
		}
	}

	// TUGAS 2: Update PAID -> EXPIRED (Past Flight)
	err = s.repo.UpdatePastBookingsToExpired()
	if err != nil {
		log.Printf("[CRON] Error updating past flights: %v\n", err)
	}

	log.Println("[CRON] --- Job Finished ---")
	return nil
}

func (s *bookingService) GetUserBookings(userID uint) ([]MyBookingResponse, error) {
	bookings, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []MyBookingResponse
	for _, b := range bookings {
		
		// 1. Mapping Flight Detail
		var seatClass, classCode string
		if len(b.Details) > 0 {
			seatClass = b.Details[0].SeatClass
		}
		// Cari Class Code (Sub-class)
		for _, fc := range b.Flight.FlightClasses {
			if strings.EqualFold(fc.SeatClass, seatClass) {
				classCode = fc.ClassCode
				break
			}
		}

		flightDetail := BookingFlightDetail{
			FlightCode:        b.Flight.FlightCode,
			AirlineName:       b.Flight.Airline.Name,
			AirlineLogo:       b.Flight.Airline.LogoURL,
			Origin:            fmt.Sprintf("%s (%s)", b.Flight.OriginAirport.CityName, b.Flight.OriginAirport.Code),
			Destination:       fmt.Sprintf("%s (%s)", b.Flight.DestinationAirport.CityName, b.Flight.DestinationAirport.Code),
			DepartureTime:     b.Flight.DepartureTime,
			ArrivalTime:       b.Flight.ArrivalTime,
			
			// [UPDATED] Ambil data langsung dari Model Flight & Utils
			DurationMinutes:   b.Flight.TotalDuration, 
			DurationFormatted: utils.FormatDuration(b.Flight.TotalDuration),
			TransitInfo:       b.Flight.TransitInfo, // Ambil dari DB ("Direct", "1 Transit")
			
			SeatClass:         seatClass,
			ClassCode:         classCode,
		}

		// 2. Mapping Passengers (NEW)
		var passengerList []PassengerDetailResponse
		for _, detail := range b.Details {
			passengerList = append(passengerList, PassengerDetailResponse{
				FullName:     detail.PassengerName,
				Type:         detail.PassengerType, // adult/child/infant
				TicketNumber: detail.TicketNumber,
				SeatClass:    detail.SeatClass,
			})
		}

		// 3. Expiry Logic
		var expiryTime *time.Time
		if b.Status == models.BookingStatusPending {
			expiryTime = b.ExpiredAt
		}

		resp := MyBookingResponse{
			OrderID:     b.OrderID,
			BookingCode: b.BookingCode,
			Status:      b.Status,
			TotalAmount: b.TotalPrice,
			CreatedAt:   b.CreatedAt,
			ExpiryTime:  expiryTime,
			Flight:      flightDetail,
			Passengers:  passengerList, // Masukkan list penumpang
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *bookingService) DownloadInvoice(ctx context.Context, bookingCode string) ([]byte, error) {
	// 1. Ambil Data Lengkap dari Repo
	// Pastikan kamu sudah menambahkan method GetBookingForInvoice di repository.go!
	booking, err := s.repo.GetBookingForInvoice(bookingCode)
	if err != nil {
		return nil, err
	}
	// 2. AMBIL DATA PAYMENT (TERPISAH)
	payment, err := s.repo.GetPaymentByOrderID(booking.OrderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	// 2. Setup Path Assets
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("gagal get working directory: %v", err)
	}
	// Path: internal/assets/images/
	assetsPath := filepath.Join(cwd, "internal", "assets", "images")
	
	getHeader := func(name string) string {
		fullPath := filepath.Join(assetsPath, name)
		b64, err := pdfprinter.ImageToBase64(fullPath)
		if err != nil {
			fmt.Printf("⚠️ Warning: Gagal load asset %s: %v\n", name, err)
			return ""
		}
		return b64
	}

	// 3. LOGIC GROUPING ITEM BELANJA (INVOICE ITEMS)
	// Kita map berdasarkan PassengerType ("DEWASA", "ANAK", "BAYI")
	type groupItem struct {
		Count int64
		Total decimal.Decimal
		Price decimal.Decimal
	}
	groupedItems := make(map[string]*groupItem)

	// Loop semua penumpang di booking_details
	for _, detail := range booking.Details {
		pType := strings.ToUpper(detail.PassengerType) // Normalisasi ke Uppercase
		
		if _, exists := groupedItems[pType]; !exists {
			// Inisialisasi
			groupedItems[pType] = &groupItem{
				Count: 0, 
				Total: decimal.Zero, 
				Price: detail.Price, // Asumsi harga per tipe sama dalam satu booking
			}
		}
		
		groupedItems[pType].Count++
		groupedItems[pType].Total = groupedItems[pType].Total.Add(detail.Price)
	}

	// Convert Map ke Struct InvoiceItem PDF
	var invoiceItems []pdfprinter.InvoiceItem
	counter := 1
	
	// Ambil Info Flight Umum (Dari Leg Pertama)
	flightDesc := "Penerbangan"
	// Cek apakah data flight terload (Eager Loading)
	if booking.Flight.ID != 0 && len(booking.Flight.FlightLegs) > 0 {
		leg := booking.Flight.FlightLegs[0]
		// Contoh: "Lion Air CGK-DPS"
		flightDesc = fmt.Sprintf("%s %s-%s", 
			leg.Airline.Name, 
			leg.OriginAirport.Code, 
			leg.DestinationAirport.Code,
		)
	} else {
		flightDesc = "Tiket Penerbangan"
	}
	
	// Tanggal Terbang (Format: 5 Nov 2025)
	var flightDateStr string
	if booking.Flight.ID != 0 {
		flightDateStr = booking.Flight.DepartureTime.Format("02 Jan 2006")
	} else {
		flightDateStr = "-"
	}

	// Loop Map Group tadi untuk jadi baris Invoice
	for pType, data := range groupedItems {
		// Description: "Lion Air CGK-DPS (DEWASA) 5 Nov 2025"
		fullDesc := fmt.Sprintf("%s (%s) %s", flightDesc, pType, flightDateStr)

		// Konversi Decimal ke Float64 untuk Formatter
		totalFloat, _ := data.Total.Float64()

		item := pdfprinter.InvoiceItem{
			Number:      strconv.Itoa(counter),
			Product:     "Tiket Pesawat",
			Description: fullDesc,
			Quantity:    int(data.Count),
			Total:       utils.FormatRupiah(totalFloat), // Pakai helper baru
		}
		invoiceItems = append(invoiceItems, item)
		counter++
	}

	// 4. Mapping Data Penumpang (Untuk List Penumpang di Invoice)
	var passengers []pdfprinter.Passenger
	for _, detail := range booking.Details {
		passengers = append(passengers, pdfprinter.Passenger{
			Name: detail.PassengerName,
			Type: fmt.Sprintf("(%s)", strings.ToUpper(detail.PassengerType)),
		})
	}

	// 5. Mapping Metadata Invoice & Payment
	paymentMethod := "Menunggu Pembayaran"
	paymentStatus := "BELUM LUNAS"
	paymentDate := "-"

	// Cek Relasi Payment (Pointer check)
	if payment != nil {
		// FIX: Mapping Nama Metode Pembayaran dari Field Midtrans
		switch payment.PaymentType {
		case "bank_transfer":
			// Jika bank transfer, ambil nama bank-nya (bca, bni, dll)
			paymentMethod = fmt.Sprintf("%s Virtual Account", strings.ToUpper(payment.Bank))
		case "echannel":
			paymentMethod = "Mandiri Bill"
		case "gopay", "qris":
			paymentMethod = "QRIS / GoPay"
		case "credit_card":
			paymentMethod = "Credit Card"
		default:
			// Fallback: rapikan string (ex: cstore -> Cstore)
			paymentMethod = strings.Title(strings.ReplaceAll(payment.PaymentType, "_", " "))
		}

		// FIX: Mapping Status dari 'transaction_status'
		// Midtrans pakai "settlement" untuk Lunas
		switch payment.TransactionStatus {
		case models.PaymentStatusSettlement: // "settlement"
			paymentStatus = "LUNAS"
		case models.PaymentStatusPending:    // "pending"
			paymentStatus = "PENDING"
		case models.PaymentStatusExpire:     // "expire"
			paymentStatus = "KADALUARSA"
		case models.PaymentStatusCancel:     // "cancel"
			paymentStatus = "DIBATALKAN"
		case models.PaymentStatusDeny:       // "deny"
			paymentStatus = "DITOLAK"
		default:
			// Fallback: Jika ada status lain, cukup di-uppercase
			paymentStatus = strings.ToUpper(payment.TransactionStatus)
		}
		
		// Tanggal Bayar (created_at payment)
		paymentDate = payment.CreatedAt.Format("02 January 2006, 15:04")
	}

	// Helper Decimal to Float
	totalAmountFloat, _ := booking.TotalPrice.Float64() // Note: Di modelmu namanya TotalPrice, bukan TotalAmount

	// 6. Susun Data Final
	invoiceData := pdfprinter.InvoiceData{
		HeaderImage:   getHeader("invoice_header.png"),
		FooterImage:   getHeader("invoice_footer.png"),
		
		InvoiceNumber: booking.BookingCode, // Menggunakan Kode Booking sbg No Invoice
		Date:          paymentDate,         // Tanggal bayar
		
		CustomerName:  booking.User.FullName,
		CustomerEmail: booking.User.Email,
		CustomerPhone: booking.User.Phone,
		
		PaymentMethod: paymentMethod,
		PaymentStatus: paymentStatus,
		
		Passengers:    passengers,
		Items:         invoiceItems,
		
		SubTotal:      utils.FormatRupiah(totalAmountFloat),
		ServiceFee:    "Rp 0", // HARDCODED 0 Sesuai Request
		GrandTotal:    utils.FormatRupiah(totalAmountFloat),
	}

	// 7. Panggil Printer Engine
	// Generate file temp unik
	tmpFileName := fmt.Sprintf("temp_invoice_%s_%d.pdf", bookingCode, time.Now().Unix())
	tmpFilePath := filepath.Join(cwd, tmpFileName)
	
	// Generate PDF
	// Pastikan const TemplateFolder di printer.go menunjuk ke folder yang benar relative terhadap cwd
	// atau hardcode path template di sini jika perlu.
	// Asumsi: "invoice.html" ada di internal/utils/pdf_printer/templates/
	err = pdfprinter.GeneratePDF("invoice.html", invoiceData, tmpFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed generate pdf: %w", err)
	}
	
	// Baca file PDF jadi bytes
	pdfBytes, err := os.ReadFile(tmpFilePath)
	if err != nil {
		os.Remove(tmpFilePath) // Cleanup if read fails
		return nil, fmt.Errorf("failed read pdf file: %w", err)
	}
	
	// Hapus file temp (Cleanup) agar server tidak penuh
	// Gunakan goroutine atau defer sebenarnya lebih aman, tapi ini oke
	os.Remove(tmpFilePath)

	return pdfBytes, nil
}

func (s *bookingService) DownloadEticket(ctx context.Context, bookingCode string) ([]byte, error) {
	// 1. AMBIL DATA BOOKING
	booking, err := s.repo.GetBookingForTicket(bookingCode)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	// 2. SETUP PATH ASSETS (SAMA SEPERTI INVOICE)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("gagal get working directory: %v", err)
	}
	assetsPath := filepath.Join(cwd, "internal", "assets", "images")

	// Helper Inline (SAMA SEPERTI INVOICE)
	getHeader := func(name string) string {
		fullPath := filepath.Join(assetsPath, name)
		b64, err := pdfprinter.ImageToBase64(fullPath)
		if err != nil {
			fmt.Printf("⚠️ Warning: Gagal load asset %s: %v\n", name, err)
			return ""
		}
		return b64
	}

	// 3. GENERATE QR CODE
	qrCodeBase64, err := pdfprinter.GenerateQRCodeBase64(booking.BookingCode)
	if err != nil {
		fmt.Printf("Warning: Failed to generate QR: %v\n", err)
		qrCodeBase64 = ""
	}

	// 4. MAPPING FLIGHT SEGMENTS & TRANSIT LOGIC
	var segments []pdfprinter.FlightSegment

	// Pastikan data flight & legs tersedia
	if booking.Flight.ID != 0 && len(booking.Flight.FlightLegs) > 0 {
		legs := booking.Flight.FlightLegs
		totalLegs := len(legs)

		for i, leg := range legs {
			// Setup Logo Airline
			// TODO: Jika nanti ada URL gambar di DB, ganti ini. Sekarang pakai placeholder.
			airlineLogo := getHeader("Lion.png") 

			// Jika di database ada URL gambar, coba download
			// Pastikan field di model Airline kamu bernama "Image" (sesuaikan jika namanya "Logo" atau "ImageUrl")
			if leg.Airline != nil && leg.Airline.LogoURL != "" {
				remoteB64, err := downloadImageToBase64(leg.Airline.LogoURL)
				if err == nil && remoteB64 != "" {
					airlineLogo = remoteB64
				} else {
					fmt.Printf("⚠️ Gagal download logo maskapai (%s): %v. Menggunakan default.\n", leg.Airline.LogoURL, err)
				}
			}

			// Hitung Durasi Terbang (Arrival - Departure)
			durationMinutes := int(leg.ArrivalTime.Sub(leg.DepartureTime).Minutes())
			durationStr := utils.FormatDuration(durationMinutes)

			// Logic Transit (Cek apakah ada leg berikutnya)
			var transitDetail pdfprinter.TransitDetail
			if i < totalLegs-1 {
				// Bukan leg terakhir, berarti TRANSIT
				nextLeg := legs[i+1]
				// Durasi Transit = Depature Next Leg - Arrival Current Leg
				transitMinutes := int(nextLeg.DepartureTime.Sub(leg.ArrivalTime).Minutes())
				
				transitDetail = pdfprinter.TransitDetail{
					IsTransit: true,
					Location:  leg.DestinationAirport.CityName, // Transit di kota tujuan saat ini
					Duration:  utils.FormatDuration(transitMinutes),
				}
			} else {
				// Leg Terakhir -> TIDAK TRANSIT
				transitDetail = pdfprinter.TransitDetail{IsTransit: false}
			}

			// Ambil Kelas Kursi (Dari Booking Details Penumpang Pertama)
			seatClass := "Economy"
			if len(booking.Details) > 0 {
				seatClass = booking.Details[0].SeatClass
			}

			// Append Segment
			segments = append(segments, pdfprinter.FlightSegment{
				AirlineName:  leg.Airline.Name,
				AirlineLogo:  airlineLogo,
				FlightNumber: booking.Flight.FlightCode, // Gunakan Flight Code utama
				FlightClass:  seatClass,
				
				Departure: pdfprinter.FlightPoint{
					Date:        leg.DepartureTime.Format("02 Jan 2006"),
					Time:        leg.DepartureTime.Format("15:04"),
					CityName:    leg.OriginAirport.CityName,
					CityCode:    leg.OriginAirport.Code,
					AirportName: leg.OriginAirport.AirportName,
				},
				Arrival: pdfprinter.FlightPoint{
					Date:        leg.ArrivalTime.Format("02 Jan 2006"),
					Time:        leg.ArrivalTime.Format("15:04"),
					CityName:    leg.DestinationAirport.CityName,
					CityCode:    leg.DestinationAirport.Code,
					AirportName: leg.DestinationAirport.AirportName,
				},
				Duration: durationStr,
				Transit:  transitDetail,
			})
		}
	}

	// 5. MAPPING PASSENGER LIST
	var passengers []pdfprinter.TicketPassenger
	for i, detail := range booking.Details {
		passengers = append(passengers, pdfprinter.TicketPassenger{
			Number:       i + 1,
			Name:         fmt.Sprintf("%s. %s (%s)", strings.ToUpper(detail.PassengerTitle), detail.PassengerName, strings.ToUpper(detail.PassengerType)),
			TicketNumber: detail.TicketNumber,
		})
	}

	// 6. SUSUN DATA FINAL
	ticketData := pdfprinter.TicketData{
		HeaderImage: getHeader("eticket_header.png"), 
		FooterImage: getHeader("eticket_footer.png"),
		
		BookingID:   booking.OrderID,
		BookingCode: booking.BookingCode,
		BookingDate: booking.CreatedAt.Format("02 Jan 2006, 15:04"),
		BookerName:  booking.User.FullName,
		QRCode:      qrCodeBase64,
		
		Segments:    segments,
		Passengers:  passengers,
	}

	// 7. GENERATE PDF (SAMA PERSIS SEPERTI INVOICE)
	tmpFileName := fmt.Sprintf("temp_ticket_%s_%d.pdf", bookingCode, time.Now().Unix())
	tmpFilePath := filepath.Join(cwd, tmpFileName)

	// Panggil Printer
	err = pdfprinter.GeneratePDF("ticket.html", ticketData, tmpFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed generate pdf: %w", err)
	}

	// Baca File
	pdfBytes, err := os.ReadFile(tmpFilePath)
	if err != nil {
		os.Remove(tmpFilePath) // Cleanup jika gagal baca
		return nil, fmt.Errorf("failed read pdf file: %w", err)
	}

	// Hapus File Temp
	os.Remove(tmpFilePath)

	return pdfBytes, nil
}

// Utils (Tetap Sama)
func generateRandomString(n int) string {
	const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func generatePNR() string {
	return fmt.Sprintf("%s", generateRandomString(6))
}

func calculatePassengerType(dob time.Time) string {
	now := time.Now()
	age := now.Year() - dob.Year()
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

func formatPaymentMethodName(code string) string {
	code = strings.ToLower(code)
	switch code {
	case "va_bca":
		return "BCA Virtual Account"
	case "va_mandiri":
		return "Mandiri Virtual Account"
	case "va_bri":
		return "BRI Virtual Account"
	case "ewallet_gopay":
		return "GoPay"
	case "ewallet_ovo":
		return "OVO"
	case "credit_card":
		return "Credit Card"
	default:
		// Fallback: ganti underscore dengan spasi dan uppercase
		return strings.ToUpper(strings.ReplaceAll(code, "_", " "))
	}
}

func downloadImageToBase64(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(body), nil
}