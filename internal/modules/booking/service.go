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
	authService   auth.AuthService
}

func NewBookingService(
	repo BookingRepository,
	flightService flight.FlightService,
	authService auth.AuthService,
) BookingService {
	return &bookingService{
		repo:          repo,
		flightService: flightService,
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

	expiryDuration := 55 * time.Minute
	expiryAt := time.Now().Add(expiryDuration)

	for _, item := range req.Items {
		flightData, err := s.flightService.GetFlightByID(item.FlightID)
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
	
	return &BookingResponse{
		OrderID:         orderID,
		TotalAmount:     grandTotal,
		Status:          models.BookingStatusPending,
		TransactionTime: time.Now(),
		ExpiryTime:      &expiryAt,
		Bookings:        bookingResponses,
	}, nil
}

func (s *bookingService) ProcessExpiredBookings() error {
	log.Println("[CRON] --- Starting Scheduler Job ---")

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
		var seatClass, classCode string
		if len(b.Details) > 0 {
			seatClass = b.Details[0].SeatClass
		}
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
			DurationMinutes:   b.Flight.TotalDuration, 
			DurationFormatted: utils.FormatDuration(b.Flight.TotalDuration),
			TransitInfo:       b.Flight.TransitInfo,
			SeatClass:         seatClass,
			ClassCode:         classCode,
		}

		var passengerList []PassengerDetailResponse
		for _, detail := range b.Details {
			passengerList = append(passengerList, PassengerDetailResponse{
				FullName:     detail.PassengerName,
				Type:         detail.PassengerType,
				TicketNumber: detail.TicketNumber,
				SeatClass:    detail.SeatClass,
			})
		}
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
			Passengers:  passengerList,
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *bookingService) DownloadInvoice(ctx context.Context, orderID string) ([]byte, error) {
	bookings, err := s.repo.GetBookingsForInvoiceByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("bookings not found: %w", err)
	}
	if len(bookings) == 0 {
		return nil, fmt.Errorf("no bookings found for order id: %s", orderID)
	}

	mainBooking := bookings[0]

	payment, err := s.repo.GetPaymentByOrderID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payment: %w", err)
	}

	cwd, _ := os.Getwd()
	assetsPath := filepath.Join(cwd, "internal", "assets", "images")
	getHeader := func(name string) string {
		fullPath := filepath.Join(assetsPath, name)
		b64, err := pdfprinter.ImageToBase64(fullPath)
		if err != nil {
			return ""
		}
		return b64
	}

	var invoiceItems []pdfprinter.InvoiceItem
	var totalAmountDecimal decimal.Decimal
	counter := 1

	for _, booking := range bookings {
		totalAmountDecimal = totalAmountDecimal.Add(booking.TotalPrice)

		type groupItem struct {
			Count int64
			Total decimal.Decimal
			Price decimal.Decimal
		}
		groupedItems := make(map[string]*groupItem)

		for _, detail := range booking.Details {
			pType := strings.ToUpper(detail.PassengerType)
			if _, exists := groupedItems[pType]; !exists {
				groupedItems[pType] = &groupItem{Count: 0, Total: decimal.Zero, Price: detail.Price}
			}
			groupedItems[pType].Count++
			groupedItems[pType].Total = groupedItems[pType].Total.Add(detail.Price)
		}

		flightDesc := "Penerbangan"
		flightDateStr := "-"
		if booking.Flight.ID != 0 {
			flightDateStr = booking.Flight.DepartureTime.Format("02 Jan 2006")
			if len(booking.Flight.FlightLegs) > 0 {
				leg := booking.Flight.FlightLegs[0]
				flightDesc = fmt.Sprintf("%s %s-%s", 
					leg.Airline.Name, 
					leg.OriginAirport.Code, 
					leg.DestinationAirport.Code,
				)
			}
		}

		for pType, data := range groupedItems {
			fullDesc := fmt.Sprintf("%s (%s) - %s", flightDesc, pType, flightDateStr)
			totalFloat, _ := data.Total.Float64()

			item := pdfprinter.InvoiceItem{
				Number:      strconv.Itoa(counter),
				Product:     "Tiket Pesawat",
				Description: fullDesc,
				Quantity:    int(data.Count),
				Total:       utils.FormatRupiah(totalFloat),
			}
			invoiceItems = append(invoiceItems, item)
			counter++
		}
	}

	var passengers []pdfprinter.Passenger
	for _, detail := range mainBooking.Details {
		passengers = append(passengers, pdfprinter.Passenger{
			Name: detail.PassengerName,
			Type: fmt.Sprintf("(%s)", strings.ToUpper(detail.PassengerType)),
		})
	}

	paymentMethod := "Menunggu Pembayaran"
	paymentStatus := "BELUM LUNAS"
	paymentDate := "-"

	if payment != nil {
		switch payment.PaymentType {
		case "bank_transfer":
			paymentMethod = fmt.Sprintf("%s Virtual Account", strings.ToUpper(payment.Bank))
		case "echannel":
			paymentMethod = "Mandiri Bill"
		case "gopay", "qris":
			paymentMethod = "QRIS / GoPay"
		case "credit_card":
			paymentMethod = "Credit Card"
		default:
			paymentMethod = strings.Title(strings.ReplaceAll(payment.PaymentType, "_", " "))
		}

		switch payment.TransactionStatus {
		case models.PaymentStatusSettlement:
			paymentStatus = "LUNAS"
		case models.PaymentStatusPending:
			paymentStatus = "PENDING"
		case models.PaymentStatusExpire:
			paymentStatus = "KADALUARSA"
		case models.PaymentStatusCancel:
			paymentStatus = "DIBATALKAN"
		case models.PaymentStatusDeny:
			paymentStatus = "DITOLAK"
		default:
			paymentStatus = strings.ToUpper(payment.TransactionStatus)
		}
		
		paymentDate = payment.CreatedAt.Format("02 January 2006, 15:04")
	}

	finalTotalFloat, _ := totalAmountDecimal.Float64()

	invoiceData := pdfprinter.InvoiceData{
		HeaderImage:   getHeader("invoice_header.png"),
		FooterImage:   getHeader("invoice_footer.png"),
		InvoiceNumber: mainBooking.OrderID,
		Date:          paymentDate,
		CustomerName:  mainBooking.User.FullName,
		CustomerEmail: mainBooking.User.Email,
		CustomerPhone: mainBooking.User.Phone,
		PaymentMethod: paymentMethod,
		PaymentStatus: paymentStatus,
		Passengers:    passengers,
		Items:         invoiceItems,
		SubTotal:      utils.FormatRupiah(finalTotalFloat),
		ServiceFee:    "Rp 0",
		GrandTotal:    utils.FormatRupiah(finalTotalFloat),
	}

	tmpFileName := fmt.Sprintf("invoice_%s_%d.pdf", orderID, time.Now().Unix())
	tmpFilePath := filepath.Join(cwd, tmpFileName)

	err = pdfprinter.GeneratePDF("invoice.html", invoiceData, tmpFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed generate pdf: %w", err)
	}

	pdfBytes, err := os.ReadFile(tmpFilePath)
	if err != nil {
		os.Remove(tmpFilePath)
		return nil, fmt.Errorf("failed read pdf file: %w", err)
	}
	os.Remove(tmpFilePath)

	return pdfBytes, nil
}

func (s *bookingService) DownloadEticket(ctx context.Context, bookingCode string) ([]byte, error) {
	booking, err := s.repo.GetBookingForTicket(bookingCode)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("gagal get working directory: %v", err)
	}
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

	qrCodeBase64, err := pdfprinter.GenerateQRCodeBase64(booking.BookingCode)
	if err != nil {
		fmt.Printf("Warning: Failed to generate QR: %v\n", err)
		qrCodeBase64 = ""
	}

	var segments []pdfprinter.FlightSegment

	if booking.Flight.ID != 0 && len(booking.Flight.FlightLegs) > 0 {
		legs := booking.Flight.FlightLegs
		totalLegs := len(legs)

		for i, leg := range legs {
			airlineLogo := getHeader("Lion.png")
			if leg.Airline != nil && leg.Airline.LogoURL != "" {
				remoteB64, err := downloadImageToBase64(leg.Airline.LogoURL)
				if err == nil && remoteB64 != "" {
					airlineLogo = remoteB64
				} else {
					fmt.Printf("⚠️ Gagal download logo maskapai (%s): %v. Menggunakan default.\n", leg.Airline.LogoURL, err)
				}
			}

			durationMinutes := int(leg.ArrivalTime.Sub(leg.DepartureTime).Minutes())
			durationStr := utils.FormatDuration(durationMinutes)

			var transitDetail pdfprinter.TransitDetail
			if i < totalLegs-1 {
				nextLeg := legs[i+1]
				transitMinutes := int(nextLeg.DepartureTime.Sub(leg.ArrivalTime).Minutes())
				
				transitDetail = pdfprinter.TransitDetail{
					IsTransit: true,
					Location:  leg.DestinationAirport.CityName,
					Duration:  utils.FormatDuration(transitMinutes),
				}
			} else {
				transitDetail = pdfprinter.TransitDetail{IsTransit: false}
			}

			seatClass := "Economy"
			if len(booking.Details) > 0 {
				seatClass = booking.Details[0].SeatClass
			}

			segments = append(segments, pdfprinter.FlightSegment{
				AirlineName:  leg.Airline.Name,
				AirlineLogo:  airlineLogo,
				FlightNumber: booking.Flight.FlightCode,
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

	var passengers []pdfprinter.TicketPassenger
	for i, detail := range booking.Details {
		passengers = append(passengers, pdfprinter.TicketPassenger{
			Number:       i + 1,
			Name:         fmt.Sprintf("%s. %s (%s)", strings.ToUpper(detail.PassengerTitle), detail.PassengerName, strings.ToUpper(detail.PassengerType)),
			TicketNumber: detail.TicketNumber,
		})
	}

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

	tmpFileName := fmt.Sprintf("temp_ticket_%s_%d.pdf", bookingCode, time.Now().Unix())
	tmpFilePath := filepath.Join(cwd, tmpFileName)

	err = pdfprinter.GeneratePDF("ticket.html", ticketData, tmpFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed generate pdf: %w", err)
	}

	pdfBytes, err := os.ReadFile(tmpFilePath)
	if err != nil {
		os.Remove(tmpFilePath)
		return nil, fmt.Errorf("failed read pdf file: %w", err)
	}

	os.Remove(tmpFilePath)

	return pdfBytes, nil
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