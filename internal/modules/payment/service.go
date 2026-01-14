package payment

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"ezytix-be/internal/config"
	"ezytix-be/internal/models"

	"github.com/xendit/xendit-go/v6"
	"github.com/xendit/xendit-go/v6/payment_request"
)

// ==========================================
// 1. DEFINISI INTERFACE
// ==========================================

// BookingDataAccessor: Hanya ambil data dan update status.
// TIDAK ADA fungsi update expiry, karena Expiry Booking bersifat STRICT/IMMUTABLE.
type BookingDataAccessor interface {
	UpdateBookingStatus(orderID string, status string) error
	GetBookingByOrderID(orderID string) (*models.Booking, error)
}

type PaymentService interface {
	InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error)
	ProcessWebhook(payload map[string]interface{}, webhookToken string) error
	FindPaymentByOrderID(orderID string) (*models.Payment, error)
}

type paymentService struct {
	repo         PaymentRepository
	bookingRepo  BookingDataAccessor
	xenditClient *xendit.APIClient
}

func NewPaymentService(repo PaymentRepository, bookingRepo BookingDataAccessor) PaymentService {
	client := xendit.NewClient(config.AppConfig.XenditSecretKey)
	return &paymentService{
		repo:         repo,
		bookingRepo:  bookingRepo,
		xenditClient: client,
	}
}

// ==========================================
// 2. INITIATE PAYMENT (BRAIN OF THE OPERATION)
// ==========================================
func (s *paymentService) InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error) {
	// 1. Ambil Data Booking
	booking, err := s.bookingRepo.GetBookingByOrderID(req.OrderID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	// 2. [STRICT VALIDATION] Cek Expiry Booking
	if booking.ExpiredAt == nil {
		return nil, errors.New("booking expiry data is corrupt")
	}
	
	timeRemaining := time.Until(*booking.ExpiredAt)
	
	// A. Jika sudah expired
	if timeRemaining <= 0 {
		return nil, errors.New("booking has expired, please create a new order")
	}

	// B. Jika waktu terlalu mepet (< 5 menit), tolak demi keamanan
	if timeRemaining < 5*time.Minute {
		return nil, errors.New("payment window closing soon (less than 5 mins), please create a new order")
	}

	// 3. Cek Idempotency (Existing Payment)
	existingPayment, _ := s.repo.FindPaymentByOrderID(req.OrderID)
	if existingPayment != nil && existingPayment.PaymentStatus == models.PaymentStatusPending {
		// Jika metode sama, return data lama
		if existingPayment.PaymentMethod == req.PaymentMethod {
			return &InitiatePaymentResponse{
				OrderID:       existingPayment.OrderID,
				PaymentMethod: existingPayment.PaymentMethod,
				PaymentCode:   existingPayment.PaymentCode,
				QrString:      existingPayment.QrString,
				DeepLink:      existingPayment.PaymentUrl,
				Amount:        booking.TotalPrice.InexactFloat64(),
				ExpiryTime:    *existingPayment.ExpiryTime, // Tetap gunakan expiry lama (karena sama dengan booking)
				Status:        models.PaymentStatusPending,
			}, nil
		}
		// Jika metode beda, kita akan overwrite di DB (Last Win Strategy)
	}

	// 4. Routing ke Xendit API
	// Kita passing Expiry Booking ke Xendit agar VA mati bersamaan dengan Booking
	bookingExpiry := *booking.ExpiredAt
	amountFloat := booking.TotalPrice.InexactFloat64()
	var resp *InitiatePaymentResponse

	switch req.PaymentType {
	case "VIRTUAL_ACCOUNT":
		resp, err = s.createFixedVA(req, amountFloat, bookingExpiry)
	case "QR_CODE":
		// QRIS Xendit V6 mungkin tidak support set expiry spesifik per request.
		// Jadi kita hanya passing untuk kebutuhan struct response backend.
		resp, err = s.createQRCode(req, amountFloat, bookingExpiry)
	default:
		return nil, errors.New("payment type not supported")
	}

	if err != nil {
		return nil, err
	}

	// 5. Simpan Payment ke Database
	paymentModel := &models.Payment{
		OrderID:        req.OrderID,
		XenditID:       "PR-" + req.OrderID, // Bisa diganti ID asli dari resp Xendit jika perlu
		PaymentMethod:  req.PaymentMethod,
		PaymentChannel: req.PaymentType,
		PaymentStatus:  models.PaymentStatusPending,
		Amount:         booking.TotalPrice,
		PaymentCode:    resp.PaymentCode,
		QrString:       resp.QrString,
		PaymentUrl:     resp.DeepLink,
		ExpiryTime:     &bookingExpiry, // Payment Expiry = Booking Expiry
	}

	// Upsert Logic
	if err := s.repo.CreatePayment(paymentModel); err != nil {
		return nil, err
	}

	return resp, nil
}

// --- HELPER: Create Virtual Account ---
func (s *paymentService) createFixedVA(req InitiatePaymentRequest, amount float64, expiryTime time.Time) (*InitiatePaymentResponse, error) {
	currency := payment_request.PAYMENTREQUESTCURRENCY_IDR

	// A. Main Params
	pr := *payment_request.NewPaymentRequestParameters(currency)
	pr.SetAmount(amount)
	pr.SetReferenceId(req.OrderID)

	// B. Channel Properties
	channelProps := *payment_request.NewVirtualAccountChannelProperties("Ezytix Customer")
	
	// [STRICT] Set Expiry VA SAMA PERSIS dengan Booking Expiry
	channelProps.SetExpiresAt(expiryTime)

	// C. Channel Code
	channelCode := payment_request.VirtualAccountChannelCode(req.PaymentMethod)

	// D. VA Params
	vaParams := *payment_request.NewVirtualAccountParameters(channelCode, channelProps)

	// E. Payment Method Params
	pmParams := *payment_request.NewPaymentMethodParameters(
		payment_request.PAYMENTMETHODTYPE_VIRTUAL_ACCOUNT,
		payment_request.PAYMENTMETHODREUSABILITY_ONE_TIME_USE,
	)
	pmParams.SetVirtualAccount(vaParams)

	pr.SetPaymentMethod(pmParams)

	// F. Execute
	resp, _, err := s.xenditClient.PaymentRequestApi.CreatePaymentRequest(context.Background()).
		PaymentRequestParameters(pr).
		Execute()

	if err != nil {
		log.Printf("Xendit VA Error: %v", err)
		return nil, errors.New("failed to create virtual account")
	}

	// G. Parse Response
	var vaNumber string
	if resp.PaymentMethod.VirtualAccount.IsSet() {
		vaData := resp.PaymentMethod.VirtualAccount.Get()
		if vaData != nil && vaData.ChannelProperties.VirtualAccountNumber != nil {
			vaNumber = *vaData.ChannelProperties.VirtualAccountNumber
		}
	}

	if vaNumber == "" {
		return nil, errors.New("xendit did not return VA number")
	}

	return &InitiatePaymentResponse{
		OrderID:       req.OrderID,
		PaymentMethod: req.PaymentMethod,
		PaymentCode:   vaNumber,
		Amount:        amount,
		ExpiryTime:    expiryTime,
		Status:        models.PaymentStatusPending,
	}, nil
}

// --- HELPER: Create QR Code ---
func (s *paymentService) createQRCode(req InitiatePaymentRequest, amount float64, expiryTime time.Time) (*InitiatePaymentResponse, error) {
	currency := payment_request.PAYMENTREQUESTCURRENCY_IDR

	// A. Main Params
	pr := *payment_request.NewPaymentRequestParameters(currency)
	pr.SetAmount(amount)
	pr.SetReferenceId(req.OrderID)

	// B. QR Params
	qrParams := *payment_request.NewQRCodeParameters()
	qrParams.SetChannelCode("QRIS")
	
	// Note: Xendit QRIS V6 tidak selalu support set expiry di level ini. 
	// Jika tidak support, QRIS akan default 30 menit atau sesuai setting merchant.
	// Namun Backend kita tetap menolak pembayaran jika booking expired.

	// C. Payment Method Params
	pmParams := *payment_request.NewPaymentMethodParameters(
		payment_request.PAYMENTMETHODTYPE_QR_CODE,
		payment_request.PAYMENTMETHODREUSABILITY_ONE_TIME_USE,
	)
	pmParams.SetQrCode(qrParams)

	pr.SetPaymentMethod(pmParams)

	// D. Execute
	resp, _, err := s.xenditClient.PaymentRequestApi.CreatePaymentRequest(context.Background()).
		PaymentRequestParameters(pr).
		Execute()

	if err != nil {
		log.Printf("Xendit QR Error: %v", err)
		return nil, errors.New("failed to create qr code")
	}

	// E. Parse Response
	var qrString string
	if resp.PaymentMethod.QrCode.IsSet() {
		qrData := resp.PaymentMethod.QrCode.Get()
		if qrData != nil && qrData.ChannelProperties.QrString != nil {
			qrString = *qrData.ChannelProperties.QrString
		}
	}

	if qrString == "" {
		return nil, errors.New("xendit did not return QR string")
	}

	return &InitiatePaymentResponse{
		OrderID:       req.OrderID,
		PaymentMethod: "QRIS",
		QrString:      qrString,
		Amount:        amount,
		ExpiryTime:    expiryTime,
		Status:        models.PaymentStatusPending,
	}, nil
}

// ==========================================
// 3. PROCESS WEBHOOK
// ==========================================
func (s *paymentService) ProcessWebhook(payload map[string]interface{}, webhookToken string) error {
	// A. Security Check
	if webhookToken != config.AppConfig.XenditWebhookToken {
		return errors.New("unauthorized webhook token")
	}

	// B. Parse Event
	eventType, ok := payload["event"].(string)
	if !ok {
		return nil 
	}
	
	// C. Filter Event: Hanya peduli Payment Succeeded
	// Jika Expired/Failed, kita biarkan saja (karena Booking otomatis expired oleh scheduler)
	if eventType != "payment.succeeded" {
		return nil 
	}

	// D. Extract Data
	data, _ := payload["data"].(map[string]interface{})
	
	referenceID, _ := data["reference_id"].(string) 
	
	if referenceID == "" {
		// Fallback Xendit ID
		xenditID, _ := data["id"].(string)
		payment, err := s.repo.FindPaymentByXenditID(xenditID)
		if err == nil {
			referenceID = payment.OrderID
		} else {
			return errors.New("unknown transaction")
		}
	}

	// E. Update Status (Gatekeeper)
	// Kita tidak perlu cek expiry di sini secara manual, karena BookingRepo.UpdateBookingStatus
	// seharusnya menolak update jika status sudah CANCELLED/EXPIRED (Tergantung implementasi repo).
	// Tapi untuk keamanan, kita update payment dulu.
	
	now := time.Now()
	
	// 1. Update Booking Status -> PAID
	err := s.bookingRepo.UpdateBookingStatus(referenceID, models.BookingStatusPaid)
	if err != nil {
		// Jika Booking sudah dicancel Scheduler, kita tandai payment FAILED (Money In, No Ticket)
		if strings.Contains(strings.ToLower(err.Error()), "cancelled") || strings.Contains(strings.ToLower(err.Error()), "expired") {
			_ = s.repo.UpdatePaymentStatus(referenceID, models.PaymentStatusFailed, &now)
			return nil
		}
		return err
	}

	// 2. Update Payment Status -> PAID
	err = s.repo.UpdatePaymentStatus(referenceID, models.PaymentStatusPaid, &now)
	
	return err
}

func (s *paymentService) FindPaymentByOrderID(orderID string) (*models.Payment, error) {
	return s.repo.FindPaymentByOrderID(orderID)
}