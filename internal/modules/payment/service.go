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

type PaymentService interface {
	CreatePayment(req CreatePaymentRequest) (*PaymentResponse, error)
	ProcessWebhook(req XenditWebhookRequest, webhookToken string) error
}

type paymentService struct {
	repo         PaymentRepository
	xenditClient *xendit.APIClient
}

func NewPaymentService(repo PaymentRepository) PaymentService {
	// Setup Xendit Client menggunakan API Key dari Config
	client := xendit.NewClient(config.AppConfig.XenditSecretKey)

	return &paymentService{
		repo:         repo,
		xenditClient: client,
	}
}

// ==========================================
// 1. CREATE PAYMENT (Call Xendit)
// ==========================================
func (s *paymentService) CreatePayment(req CreatePaymentRequest) (*PaymentResponse, error) {
	// A. Siapkan Request ke Xendit
	createInvoiceRequest := *invoice.NewCreateInvoiceRequest(
		req.OrderID, // External ID (Order ID Kita)
		req.Amount,
	)
	
	// Tambahkan data opsional agar dashboard Xendit rapi
	createInvoiceRequest.SetPayerEmail(req.PayerEmail)
	createInvoiceRequest.SetDescription(req.Description)
	
	// Set durasi expired (misal 1 jam / 3600 detik)
	// Xendit defaultnya 24 jam, kita override jadi 1 jam sesuai flow tiket pesawat
	createInvoiceRequest.SetInvoiceDuration("3600") 

	// B. Tembak API Xendit (Create Invoice)
	resp, _, err := s.xenditClient.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(createInvoiceRequest).
		Execute()

	if err != nil {
		return nil, errors.New("gagal membuat invoice xendit: " + err.Error())
	}

	// C. Simpan Response Xendit ke Database Kita (Tabel Payments)
	// Kita simpan status awal 'PENDING'
	paymentModel := &models.Payment{
		OrderID:       req.OrderID,
		XenditID:      *resp.Id,
		PaymentMethod: "INVOICE_XENDIT",
		PaymentStatus: models.PaymentStatusPending,
		Amount:        req.Amount,
		PaymentURL:    resp.InvoiceUrl, // Ini Link Pembayarannya!
	}

	if err := s.repo.CreatePayment(paymentModel); err != nil {
		return nil, err
	}

	// D. Return Data ke Booking Service (untuk dikirim ke Frontend)
	// Ambil Expiry Date dari response Xendit (time string -> time.Time)
	// Xendit return format ISO8601 usually
	
	return &PaymentResponse{
		OrderID:    req.OrderID,
		XenditID:   *resp.Id,
		PaymentURL: resp.InvoiceUrl,
		Status:     models.PaymentStatusPending,
		// ExpiryDate bisa diparsing dari resp.ExpiryDate jika perlu
	}, nil
}

// ==========================================
// 2. PROCESS WEBHOOK (Callback Logic)
// ==========================================
func (s *paymentService) ProcessWebhook(req XenditWebhookRequest, webhookToken string) error {
	// A. Security Check (Verifikasi Token)
	// Pastikan yang nembak endpoint ini benar-benar Xendit
	if webhookToken != config.AppConfig.XenditWebhookToken {
		return errors.New("unauthorized webhook token")
	}

	// B. Cari Payment di Database berdasarkan Xendit ID
	// Kita pakai Xendit ID dari JSON Callback
	payment, err := s.repo.FindPaymentByXenditID(req.ID)
	if err != nil {
		return errors.New("payment not found for xendit id: " + req.ID)
	}

	// C. Cek Status (Idempotency)
	// Kalau di database sudah PAID, jangan diproses lagi (biar stok gak error)
	if payment.PaymentStatus == models.PaymentStatusPaid {
		return nil // Sudah lunas, abaikan saja
	}

	// D. Tentukan Status Baru
	newStatus := models.PaymentStatusPending
	var paidAt *time.Time

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

	// TODO: Nanti di sini kita akan panggil Booking Repo 
	// untuk update status Booking jadi PAID juga.
	// Tapi karena Booking Repo belum ada, kita skip dulu.
	
	return nil
}