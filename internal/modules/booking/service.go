package booking

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"ezytix-be/internal/models"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/utils"

	// [REMOVED] import "ezytix-be/internal/modules/payment"

	"github.com/shopspring/decimal"
)

type BookingService interface {
	CreateOrder(userID uint, req CreateOrderRequest) (*BookingResponse, error)
	ProcessExpiredBookings() error
	GetUserBookings(userID uint) ([]MyBookingResponse, error)
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
	return fmt.Sprintf("EZY%s", generateRandomString(5))
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