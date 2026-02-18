package payment

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"ezytix-be/internal/config"
	"ezytix-be/internal/models"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

// ==========================================
// 1. DEPENDENCY CONTRACT
// ==========================================
// Agar Payment tidak bergantung langsung pada Booking (Loose Coupling)
type BookingServiceContract interface {
	GetBookingByOrderID(orderID string) (*models.Booking, error)
	UpdateBookingStatus(orderID string, status string) error
}

type PaymentService interface {
	InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error)
	ProcessWebhook(payload map[string]interface{}) error
	CancelPayment(orderID string) error // [BARU]
	GetPaymentByOrderID(orderID string) (*InitiatePaymentResponse, error) // [BARU]
}

type paymentService struct {
	repo           PaymentRepository
	bookingRepo    BookingServiceContract
	midtransClient coreapi.Client
}

func NewPaymentService(repo PaymentRepository, bookingRepo BookingServiceContract) PaymentService {
	// 1. Init Client Midtrans
	var client coreapi.Client
	
	// Tentukan Environment (Sandbox / Production)
	env := midtrans.Sandbox
	if config.AppConfig.MidtransIsProduction {
		env = midtrans.Production
	}

	client.New(config.AppConfig.MidtransServerKey, env)

	return &paymentService{
		repo:           repo,
		bookingRepo:    bookingRepo,
		midtransClient: client,
	}
}

// ==========================================
// 2. CORE LOGIC: INITIATE PAYMENT
// ==========================================

func (s *paymentService) InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error) {
	// A. Validasi Booking
	booking, err := s.bookingRepo.GetBookingByOrderID(req.OrderID)
	if err != nil {
		return nil, errors.New("booking not found")
	}
	if booking.Status == models.BookingStatusPaid {
		return nil, errors.New("booking already paid")
	}

	// [LOGIC BARU] Cek Expired berdasarkan Waktu Booking (Source of Truth)
	if booking.ExpiredAt == nil {
		return nil, errors.New("booking expiry data is invalid") // Safety check
	}

	// 1. Cek apakah Booking SUDAH Expired sebelum masuk payment gateway
	if time.Now().After(*booking.ExpiredAt) {
		return nil, errors.New("booking expired")
	}

	// B. Idempotency Check
	existing, _ := s.repo.FindPaymentByOrderID(req.OrderID)
	if existing != nil && existing.TransactionStatus == "pending" {
		if existing.PaymentType == req.PaymentType {
			return s.constructResponseFromModel(existing), nil
		}
	}

	// ====================================================
	// C. DYNAMIC EXPIRY LOGIC (Booking Oriented)
	// ====================================================

	// 1. Hitung Sisa Waktu (Remaining Duration)
	// Rumus: Waktu Expire Booking - Waktu Sekarang
	remainingDuration := booking.ExpiredAt.Sub(time.Now())
	
	// Konversi ke Menit (Pembulatan ke bawah otomatis oleh int64)
	minutesLeft := int(remainingDuration.Minutes())

	// 2. Safety Guard
	// Jika sisa waktu kurang dari 1 menit (misal user klik di detik-detik terakhir),
	// Sebaiknya tolak payment untuk mencegah Race Condition (User bayar tapi sistem keburu expire)
	if minutesLeft < 1 {
		return nil, errors.New("booking time is almost up, please re-book")
	}

	// 3. Siapkan Time Format untuk Midtrans (Required Timezone)
	now := time.Now()
	midtransTimeFormat := now.Format("2006-01-02 15:04:05 -0700")

	// D. Siapkan Parameter Midtrans
	grossAmt := int64(booking.TotalPrice.InexactFloat64())
	
	chargeReq := &coreapi.ChargeReq{
		PaymentType: coreapi.CoreapiPaymentType(req.PaymentType),
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: grossAmt,
		},
		CustomExpiry: &coreapi.CustomExpiry{
			OrderTime:      midtransTimeFormat, // Start Time = Sekarang
			ExpiryDuration: minutesLeft,        // Durasi = Sisa waktu booking
			Unit:           "minute",
		},
		Metadata: map[string]interface{}{
			"user_id": fmt.Sprintf("%d", booking.UserID),
		},
	}

	// E. Konfigurasi Spesifik per Metode (Switch Case)
	switch req.PaymentType {
	case "bank_transfer":
		chargeReq.BankTransfer = &coreapi.BankTransferDetails{
			Bank: midtrans.Bank(req.Bank),
		}
	case "echannel":
		chargeReq.EChannel = &coreapi.EChannelDetail{
			BillInfo1: "Payment For:",
			BillInfo2: "Flight Ticket",
		}
	case "qris":
		chargeReq.Qris = &coreapi.QrisDetails{
			Acquirer: "gopay",
		}
	case "gopay":
		chargeReq.Gopay = &coreapi.GopayDetails{
			EnableCallback: true,
			CallbackUrl:    config.AppConfig.FrontendURL + "/payment-finish",
		}
	default:
		return nil, errors.New("unsupported payment type")
	}

	// F. Eksekusi Charge ke Midtrans
	resp, midtransErr := s.midtransClient.ChargeTransaction(chargeReq)
	if midtransErr != nil {
		return nil, fmt.Errorf("midtrans error: %v", midtransErr)
	}

	// G. Mapping Response & Save DB
	// PENTING: Kita kirim booking.ExpiredAt (Source of Truth) ke fungsi save
	// Agar tabel 'payments' menyimpan waktu expiry yang SAMA PERSIS dengan tabel 'bookings'
	return s.saveAndRespond(booking, req, resp, booking.ExpiredAt)
}

// ==========================================
// 3. HELPERS: RESPONSE PARSING & SAVING
// ==========================================

func (s *paymentService) saveAndRespond(booking *models.Booking, req InitiatePaymentRequest, resp *coreapi.ChargeResponse, strictExpiry *time.Time) (*InitiatePaymentResponse, error) {
	
	// Variables untuk menampung hasil ekstraksi
	var vaNumber, bank, billKey, billerCode, qrUrl, deeplink string

	// 1. Ekstraksi Data (Tergantung Tipe)
	switch req.PaymentType {
	case "bank_transfer":
		if len(resp.VaNumbers) > 0 {
			vaNumber = resp.VaNumbers[0].VANumber
			bank = resp.VaNumbers[0].Bank
		} else if resp.PermataVaNumber != "" { // Permata kadang beda field
			vaNumber = resp.PermataVaNumber
			bank = "permata"
		}
	case "echannel":
		billKey = resp.BillKey
		billerCode = resp.BillerCode
	case "qris", "gopay":
		// Cari URL di Actions (standar Core API v2)
		for _, action := range resp.Actions {
			if action.Name == "generate-qr-code" {
				qrUrl = action.URL
			}
			if action.Name == "deeplink-redirect" {
				deeplink = action.URL
			}
		}
	}
	
	paymentModel := &models.Payment{
		OrderID:           req.OrderID,
		TransactionID:     resp.TransactionID,
		PaymentType:       req.PaymentType,
		GrossAmount:       booking.TotalPrice, // Pakai decimal dari booking
		TransactionStatus: resp.TransactionStatus,
		
		Bank:       bank,
		VaNumber:   vaNumber,
		BillKey:    billKey,
		BillerCode: billerCode,
		QrUrl:      qrUrl,
		Deeplink:   deeplink,
		
		ExpiryTime: strictExpiry,
	}

	if err := s.repo.CreatePayment(paymentModel); err != nil {
		return nil, err
	}

	// 3. Return DTO Response
	return s.constructResponseFromModel(paymentModel), nil
}

// Helper untuk mengubah Model DB menjadi DTO Response Frontend
func (s *paymentService) constructResponseFromModel(p *models.Payment) *InitiatePaymentResponse {
	resp := &InitiatePaymentResponse{
		OrderID:           p.OrderID,
		TransactionID:     p.TransactionID,
		PaymentType:       p.PaymentType,
		Amount:            p.GrossAmount.InexactFloat64(),
		TransactionStatus: p.TransactionStatus,
	}

	if p.ExpiryTime != nil {
		resp.ExpiryTime = *p.ExpiryTime
	}

	// Isi field dinamis berdasarkan tipe
	switch p.PaymentType {
	case "bank_transfer":
		resp.VirtualAccount = &VirtualAccountData{
			Bank:     p.Bank,
			VaNumber: p.VaNumber,
		}
	case "echannel":
		resp.MandiriBill = &MandiriBillData{
			BillKey:    p.BillKey,
			BillerCode: p.BillerCode,
		}
	case "qris":
		resp.Qris = &QrisData{
			QrUrl: p.QrUrl,
		}
	case "gopay":
		resp.Gopay = &GopayData{
			Deeplink: p.Deeplink,
		}
	}

	return resp
}

// ==========================================
// 4. CORE LOGIC: WEBHOOK PROCESSOR
// ==========================================

func (s *paymentService) ProcessWebhook(payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	transactionID, _ := payload["transaction_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signatureKey, _ := payload["signature_key"].(string)
	transactionStatus, _ := payload["transaction_status"].(string)

	// 1. Verify Signature
	input := orderID + statusCode + grossAmount + config.AppConfig.MidtransServerKey
	hash := sha512.Sum512([]byte(input))
	if signatureKey != hex.EncodeToString(hash[:]) {
		return errors.New("invalid signature key")
	}

	// 2. Map Status
	var internalStatus string
	isPaid := false
	switch transactionStatus {
	case "capture", "settlement":
		internalStatus = models.PaymentStatusSettlement
		isPaid = true
	case "pending":
		internalStatus = models.PaymentStatusPending
	case "deny", "cancel", "expire":
		internalStatus = models.PaymentStatusCancel
	default:
		internalStatus = transactionStatus
	}

	// 3. Update Payment by Transaction ID (Safe for multiple attempts)
	var paidAt *time.Time
	if isPaid {
		now := time.Now()
		paidAt = &now
	}

	if err := s.repo.UpdatePaymentStatusByTransactionID(transactionID, internalStatus, paidAt); err != nil {
		return err
	}

	// 4. Update Booking if Paid
	if isPaid {
		return s.bookingRepo.UpdateBookingStatus(orderID, models.BookingStatusPaid)
	}

	return nil
}

func (s *paymentService) CancelPayment(orderID string) error {
	// 1. Ambil transaksi terbaru untuk order ini
	payment, err := s.repo.FindPaymentByOrderID(orderID)
	if err != nil {
		return err
	}

	// 2. Perluas pengecekan status (Pending ATAU Waiting)
	isPending := payment.TransactionStatus == models.PaymentStatusPending || payment.TransactionStatus == "waiting"

	if isPending {
		// 3. HIT MIDTRANS menggunakan TransactionID (Sangat Akurat)
		fmt.Printf("🚀 Cancelling Transaction ID: %s for Order: %s\n", payment.TransactionID, orderID)
		
		_, midErr := s.midtransClient.CancelTransaction(payment.TransactionID)
		if midErr != nil {
			// Jika Midtrans bilang sudah cancel/expire (412/404), kita lanjut saja
			fmt.Printf("⚠️ Midtrans Cancel Note: %v\n", midErr)
		}

		// 4. Update status di DB lokal berdasarkan TransactionID
		// Gunakan fungsi repo yang kita buat di Step 1 sebelumnya
		return s.repo.UpdatePaymentStatusByTransactionID(payment.TransactionID, models.PaymentStatusCancel, nil)
	}

	return nil
}

func (s *paymentService) GetPaymentByOrderID(orderID string) (*InitiatePaymentResponse, error) {
	// A. Cari data payment di Database
	payment, err := s.repo.FindPaymentByOrderID(orderID)
	if err != nil {
		return nil, err
	}

	// B. Manual Mapping (Model -> DTO)
	// Kita pindahkan data dari struct 'models.Payment' ke struct response
	resp := &InitiatePaymentResponse{
		OrderID:           payment.OrderID,
		TransactionID:     payment.TransactionID,
		PaymentType:       payment.PaymentType,
		// Konversi Decimal ke Float untuk JSON number
		Amount:            payment.GrossAmount.InexactFloat64(), 
		TransactionStatus: payment.TransactionStatus,
	}

	if payment.ExpiryTime != nil {
		resp.ExpiryTime = *payment.ExpiryTime
	}

	// C. Mapping Data Spesifik per Metode Pembayaran
	switch payment.PaymentType {
	case "bank_transfer":
		// Pastikan struct response kamu punya field VirtualAccount
		resp.VirtualAccount = &VirtualAccountData{
			Bank:     payment.Bank,
			VaNumber: payment.VaNumber,
		}
	case "echannel":
		resp.MandiriBill = &MandiriBillData{
			BillKey:    payment.BillKey,
			BillerCode: payment.BillerCode,
		}
	case "qris":
		resp.Qris = &QrisData{
			QrUrl: payment.QrUrl,
		}
	case "gopay":
		resp.Gopay = &GopayData{
			Deeplink: payment.Deeplink,
		}
	}

	return resp, nil
}