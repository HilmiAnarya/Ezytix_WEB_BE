package payment

import (
	"time"
	"github.com/shopspring/decimal"
)

// ==========================================
// 1. INTERNAL DTO (Service to Service)
// ==========================================

// Request: Data yang dikirim Booking Service ke Payment Service
type CreatePaymentRequest struct {
	OrderID     string  `json:"order_id"`
	Amount      decimal.Decimal `json:"amount"`
	
	// Data user opsional untuk Invoice Xendit yang lebih rapi
	PayerName   string  `json:"payer_name"`
	PayerEmail  string  `json:"payer_email"`
	Description string  `json:"description"` // Misal: "Payment for Order #123"
}

// Response: Data yang dikembalikan Payment Service ke Booking Service
type PaymentResponse struct {
	OrderID     string    `json:"order_id"`
	XenditID    string    `json:"xendit_id"`
	PaymentURL  string    `json:"payment_url"` // Link redirect atau QR String
	Status      string    `json:"status"`
	ExpiryDate  *time.Time `json:"expiry_date"` // Kapan link ini kadaluarsa
}

// ==========================================
// 2. WEBHOOK DTO (Xendit to Backend)
// ==========================================

// Ini struktur JSON standar dari Xendit Invoice Callback
// Kita hanya ambil field yang penting saja
type XenditWebhookRequest struct {
	ID           string  `json:"id"`           		// Xendit Invoice ID
	ExternalID   string  `json:"external_id"`  		// Order ID Kita
	Status       string  `json:"status"`       		// PAID, EXPIRED, FAILED
	Amount       decimal.Decimal `json:"amount"`    // Jumlah yang dibayar
	PaidAt       string  `json:"paid_at"`      		// Waktu bayar (String ISO8601)
	
	// Field tambahan untuk validasi keamanan (opsional tapi recommended)
	Created      string  `json:"created"`
	Updated      string  `json:"updated"`
}