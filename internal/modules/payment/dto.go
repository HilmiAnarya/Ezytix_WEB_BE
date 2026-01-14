package payment

import (
	"time"
)

// ==========================================
// 1. INITIATE PAYMENT (Frontend -> Backend)
// ==========================================

// Request dari Frontend saat user klik "Bayar" setelah memilih metode
type InitiatePaymentRequest struct {
	OrderID       string `json:"order_id" validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required"` // Contoh: "BCA", "MANDIRI", "OVO", "QRIS"
	PaymentType   string `json:"payment_type" validate:"required"`   // Contoh: "VIRTUAL_ACCOUNT", "E_WALLET", "QR_CODE"
}

// Response ke Frontend (Data untuk ditampilkan di Waiting Page)
type InitiatePaymentResponse struct {
	OrderID     string    `json:"order_id"`
	PaymentMethod string  `json:"payment_method"`
	
	// Field dinamis tergantung metode bayar
	PaymentCode string    `json:"payment_code,omitempty"` // Nomor VA (untuk Virtual Account)
	QrString    string    `json:"qr_string,omitempty"`    // String QR (untuk QRIS)
	DeepLink    string    `json:"deep_link,omitempty"`    // Link redirect app (untuk E-Wallet seperti ShopeePay/Gopay)
	
	Amount      float64   `json:"amount"`
	ExpiryTime  time.Time `json:"expiry_time"`
	Status      string    `json:"status"`
}

// ==========================================
// 2. WEBHOOK DTO (Xendit -> Backend)
// ==========================================

type XenditNotification struct {
    Event string                 `json:"event"` // e.g., "payment.succeeded"
    Data  map[string]interface{} `json:"data"`  // Isinya dinamis
}