package payment

import (
	"context"
	"errors"
	"time"

	"ezytix-be/internal/config"
	"ezytix-be/internal/models"

	"github.com/xendit/xendit-go/v6"
	"github.com/xendit/xendit-go/v6/invoice"
)

// ==========================================
// 1. DEFINISI INTERFACE (CONTRACT)
// ==========================================
// Payment Service mendefinisikan apa yang dia butuhkan dari modul lain.
type BookingStatusUpdater interface {
	UpdateBookingStatus(orderID string, status string) error
}

type PaymentService interface {
	CreatePayment(req CreatePaymentRequest) (*PaymentResponse, error)
	ProcessWebhook(req XenditWebhookRequest, webhookToken string) error
}

type paymentService struct {
	repo         PaymentRepository
	bookingRepo  BookingStatusUpdater
	xenditClient *xendit.APIClient
}

func NewPaymentService(repo PaymentRepository, bookingRepo BookingStatusUpdater) PaymentService {
	client := xendit.NewClient(config.AppConfig.XenditSecretKey)

	return &paymentService{
		repo:         repo,
		bookingRepo:  bookingRepo,
		xenditClient: client,
	}
}

// ==========================================
// 1. CREATE PAYMENT (Call Xendit)
// ==========================================
func (s *paymentService) CreatePayment(req CreatePaymentRequest) (*PaymentResponse, error) {
	createInvoiceRequest := *invoice.NewCreateInvoiceRequest(req.OrderID, req.Amount)
	createInvoiceRequest.SetPayerEmail(req.PayerEmail)
	createInvoiceRequest.SetDescription(req.Description)
	createInvoiceRequest.SetInvoiceDuration("3600")

	resp, _, err := s.xenditClient.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(createInvoiceRequest).
		Execute()

	if err != nil {
		return nil, errors.New("gagal membuat invoice xendit: " + err.Error())
	}

	paymentModel := &models.Payment{
		OrderID:       req.OrderID,
		XenditID:      *resp.Id,
		PaymentMethod: "INVOICE_XENDIT",
		PaymentStatus: models.PaymentStatusPending,
		Amount:        req.Amount,
		PaymentURL:    resp.InvoiceUrl,
	}

	if err := s.repo.CreatePayment(paymentModel); err != nil {
		return nil, err
	}

	return &PaymentResponse{
		OrderID:    req.OrderID,
		XenditID:   *resp.Id,
		PaymentURL: resp.InvoiceUrl,
		Status:     models.PaymentStatusPending,
	}, nil
}

// ==========================================
// 2. PROCESS WEBHOOK (Callback Logic)
// ==========================================
func (s *paymentService) ProcessWebhook(req XenditWebhookRequest, webhookToken string) error {
	// A. Security Check
	if webhookToken != config.AppConfig.XenditWebhookToken {
		return errors.New("unauthorized webhook token")
	}

	// B. Cari Payment
	payment, err := s.repo.FindPaymentByXenditID(req.ID)
	if err != nil {
		return errors.New("payment not found for xendit id: " + req.ID)
	}

	// C. Idempotency Check
	if payment.PaymentStatus == models.PaymentStatusPaid {
		return nil 
	}

	// D. Tentukan Status Baru
	newStatus := models.PaymentStatusPending
	var paidAt *time.Time

	// Mapping Status Xendit ke Status Internal
	switch req.Status {
	case "PAID", "SETTLED":
		newStatus = models.PaymentStatusPaid
		now := time.Now()
		paidAt = &now
	case "EXPIRED":
		newStatus = models.PaymentStatusExpired
	case "FAILED":
		newStatus = models.PaymentStatusFailed
	}

	// E. Update Database Payment
	if err := s.repo.UpdatePaymentStatus(payment.OrderID, newStatus, paidAt); err != nil {
		return err
	}

	// F. [THE INTERFACE CALL] Update Booking Status
	
	bookingStatus := models.BookingStatusPending
	if newStatus == models.PaymentStatusPaid {
		bookingStatus = models.BookingStatusPaid
	} else if newStatus == models.PaymentStatusExpired {
		bookingStatus = models.BookingStatusCancelled 
	} else if newStatus == models.PaymentStatusFailed {
		bookingStatus = models.BookingStatusFailed
	}

	// Panggil Interface
	if err := s.bookingRepo.UpdateBookingStatus(payment.OrderID, bookingStatus); err != nil {
		return err 
	}
	
	return nil
}