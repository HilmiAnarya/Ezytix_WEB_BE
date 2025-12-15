package booking

import (
	"errors"
	"fmt"
	"math/rand"
	"log"
	"time"

	"ezytix-be/internal/models"
	"ezytix-be/internal/modules/auth"
	"ezytix-be/internal/modules/flight"
	"ezytix-be/internal/modules/payment"
)

type BookingService interface {
	CreateOrder(userID uint, req CreateOrderRequest) (*BookingResponse, error)
	ProcessExpiredBookings() error
}

type bookingService struct {
	repo           BookingRepository
	flightService  flight.FlightService
	paymentService payment.PaymentService
	authService    auth.AuthService
}

func NewBookingService(
	repo BookingRepository,
	flightService flight.FlightService,
	paymentService payment.PaymentService,
	authService auth.AuthService,
) BookingService {
	return &bookingService{
		repo:           repo,
		flightService:  flightService,
		paymentService: paymentService,
		authService:    authService,
	}
}

func (s *bookingService) CreateOrder(userID uint, req CreateOrderRequest) (*BookingResponse, error) {
	user, err := s.authService.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	orderID := fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), generateRandomString(4))

	var bookingsToSave []models.Booking
	var grandTotal float64
	var bookingResponses []BookingDetailResponse

	tripType := models.TripTypeOneWay
	if len(req.Items) > 1 {
		tripType = models.TripTypeRoundTrip
	}

	for _, item := range req.Items {
		flightData, err := s.flightService.GetFlightByID(item.FlightID)
		if err != nil {
			return nil, errors.New("flight not found")
		}

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

		passengerCount := len(item.Passengers)
		flightTotalPrice := selectedClass.Price * float64(passengerCount)
		grandTotal += flightTotalPrice

		bookingCode := generatePNR()
		
		booking := models.Booking{
			OrderID:         orderID,
			UserID:          userID,
			FlightID:        item.FlightID,
			BookingCode:     bookingCode,
			TripType:        tripType,
			TotalPassengers: passengerCount,
			TotalPrice:      flightTotalPrice,
			Status:          models.BookingStatusPending,
		}

		var details []models.BookingDetail
		for _, pReq := range item.Passengers {
			dobTime, _ := time.Parse("2006-01-02", pReq.DOB)
			passengerType := calculatePassengerType(dobTime)

			ticketNum := fmt.Sprintf("%s-%s", bookingCode, generateRandomString(3)) // Contoh: EZY-882X-A01

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

		originCity := flightData.OriginAirport.CityName
		destCity := flightData.DestinationAirport.CityName

		bookingResponses = append(bookingResponses, BookingDetailResponse{
			BookingCode:     bookingCode,
			FlightCode:      flightData.FlightCode,
			Origin:          originCity, 
			Destination:     destCity,   
			DepartureTime:   flightData.DepartureTime,
			TotalPassengers: passengerCount,
			TotalPrice:      flightTotalPrice,
		})
	}

	if err := s.repo.CreateOrder(bookingsToSave); err != nil {
		return nil, err
	}

	paymentReq := payment.CreatePaymentRequest{
		OrderID:     orderID,
		Amount:      grandTotal,
		Description: fmt.Sprintf("Flight Booking %s (%s)", orderID, tripType),
		PayerEmail:  user.Email, 
		PayerName:   user.FullName, 
	}

	paymentResp, err := s.paymentService.CreatePayment(paymentReq)
	if err != nil {
		return nil, errors.New("booking created but payment initiation failed: " + err.Error())
	}

	return &BookingResponse{
		OrderID:         orderID,
		TotalAmount:     grandTotal,
		Status:          models.BookingStatusPending,
		TransactionTime: time.Now(),
		PaymentURL:      paymentResp.PaymentURL, 
		Bookings:        bookingResponses,
	}, nil
}

func (s *bookingService) ProcessExpiredBookings() error {
	log.Println("[CRON] Checking for expired bookings...")

	expirationDuration := time.Minute * 2 
	expiryTime := time.Now().Add(-expirationDuration)

	expiredBookings, err := s.repo.GetExpiredBookings(expiryTime)
	if err != nil {
		log.Printf("[CRON] Error fetching expired bookings: %v\n", err)
		return err
	}

	if len(expiredBookings) == 0 {
		log.Println("[CRON] No expired bookings found.")
		return nil
	}

	log.Printf("[CRON] Found %d expired bookings. Processing cancellations...\n", len(expiredBookings))

	successCount := 0
	for _, booking := range expiredBookings {
		err := s.repo.CancelBookingAtomic(&booking)
		if err != nil {
			log.Printf("[CRON] Failed to cancel Booking ID %d (Order %s): %v\n", booking.ID, booking.OrderID, err)
			continue 
		}
		
		log.Printf("[CRON] Success cancel Booking ID %d (Order %s). Stock restored.\n", booking.ID, booking.OrderID)
		successCount++
	}

	log.Printf("[CRON] Job Finished. Successfully cancelled %d/%d bookings.\n", successCount, len(expiredBookings))
	return nil
}

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