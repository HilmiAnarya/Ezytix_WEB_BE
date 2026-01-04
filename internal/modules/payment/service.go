package payment

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	// 🔥 TAMBAHKAN BARIS INI AGAR BISA DIPANGGIL DARI LUAR 🔥
	FindPaymentByOrderID(orderID string) (*models.Payment, error)
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
	amountFloat, _ := req.Amount.Float64()

	createInvoiceRequest := *invoice.NewCreateInvoiceRequest(req.OrderID, amountFloat)
	createInvoiceRequest.SetPayerEmail(req.PayerEmail)
	createInvoiceRequest.SetDescription(req.Description)
	createInvoiceRequest.SetInvoiceDuration("3600")

	// 🔥 [STEP 2 IMPLEMENTATION] Inject Redirect URL
	// Pastikan FRONTEND_URL di .env tidak memiliki slash di akhir (misal: http://localhost:5173)
	successURL := fmt.Sprintf("%s/booking/success?order_id=%s", config.AppConfig.FrontendURL, req.OrderID)
	createInvoiceRequest.SetSuccessRedirectUrl(successURL)

	// (Opsional) Jika ingin handle redirect saat gagal
	// failureURL := fmt.Sprintf("%s/booking/failed?order_id=%s", config.AppConfig.FrontendURL, req.OrderID)
	// createInvoiceRequest.SetFailureRedirectUrl(failureURL)

	resp, _, err := s.xenditClient.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(createInvoiceRequest).
		Execute()

	if err != nil {
		// Log error detail untuk debugging jika Xendit menolak
		log.Printf("Xendit Error: %v", err)
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
// 2. PROCESS WEBHOOK (REFACTORED LOGIC)
// ==========================================
func (s *paymentService) ProcessWebhook(req XenditWebhookRequest, webhookToken string) error {
	// A. Security Check
	if webhookToken != config.AppConfig.XenditWebhookToken {
		return errors.New("unauthorized webhook token")
	}

	// B. Cari Payment (Read Only - Aman)
	payment, err := s.repo.FindPaymentByXenditID(req.ID)
	if err != nil {
		return errors.New("payment not found for xendit id: " + req.ID)
	}

	// C. Idempotency Check
	if payment.PaymentStatus == models.PaymentStatusPaid {
		return nil
	}

	// D. Tentukan Status Baru (Mapping)
	var newPaymentStatus string
	var bookingStatus string
	var paidAt *time.Time

	switch req.Status {
	case "PAID", "SETTLED":
		newPaymentStatus = models.PaymentStatusPaid
		bookingStatus = models.BookingStatusPaid
		now := time.Now()
		paidAt = &now
	case "EXPIRED":
		newPaymentStatus = models.PaymentStatusExpired
		bookingStatus = models.BookingStatusCancelled
	case "FAILED":
		newPaymentStatus = models.PaymentStatusFailed
		bookingStatus = models.BookingStatusFailed
	default:
		// Status lain dari Xendit yang tidak kita handle
		return nil
	}

	// =================================================================
	// E. LANGKAH 1: UPDATE BOOKING DULUAN (The Gatekeeper)
	// =================================================================
	// Kita coba update booking. Repo Booking akan menolak jika statusnya sudah Cancelled.
	
	err = s.bookingRepo.UpdateBookingStatus(payment.OrderID, bookingStatus)
	
	if err != nil {
		// E.1. HANDLE ZOMBIE BOOKING (Critical Problem 1)
		// Karena kita tidak bisa import variable error dari booking (Circular Dependency),
		// Kita cek pesan errornya secara manual.
		if err.Error() == "booking already cancelled by scheduler" {
			log.Printf("[CRITICAL] Race Condition Detected for Order %s. Payment received but Booking already Cancelled.", payment.OrderID)
			
			// Action: Ubah status Payment jadi FAILED agar Finance notic
			// (Uang masuk tapi tiket batal -> Perlu Refund Manual)
			_ = s.repo.UpdatePaymentStatus(payment.OrderID, models.PaymentStatusFailed, paidAt)
			
			return nil // Return nil agar Xendit tidak retry (Case closed)
		}

		// E.2. HANDLE DB ERROR (Critical Problem 2)
		// Jika errornya bukan karena Zombie (misal DB putus), kita return error.
		// Xendit akan membaca ini sebagai 500 dan melakukan RETRY nanti.
		return err 
	}

	// =================================================================
	// F. LANGKAH 2: UPDATE PAYMENT (Jika Booking Aman)
	// =================================================================
	if err := s.repo.UpdatePaymentStatus(payment.OrderID, newPaymentStatus, paidAt); err != nil {
		// Jika di sini gagal (sangat jarang), Booking sudah PAID tapi Payment masih PENDING.
		// Ini tidak fatal. Xendit akan Retry webhook.
		// Saat retry, Langkah 1 (Update Booking) akan sukses (idempotent) atau skip,
		// lalu masuk ke sini lagi untuk fix status Payment.
		return err
	}

	return nil
}

func (s *paymentService) FindPaymentByOrderID(orderID string) (*models.Payment, error) {
    return s.repo.FindPaymentByOrderID(orderID)
}