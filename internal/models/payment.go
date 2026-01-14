package models

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	PaymentStatusPending = "PENDING"
	PaymentStatusPaid    = "PAID"
	PaymentStatusExpired = "EXPIRED"
	PaymentStatusFailed  = "FAILED"
)

type Payment struct {
	ID            int             `json:"id" gorm:"primaryKey"`
	OrderID       string          `json:"order_id" gorm:"index"` // Relasi ke Booking
	
	// Xendit Info
	XenditID      string          `json:"xendit_id"`      // ID Transaksi dari Xendit
	PaymentMethod string          `json:"payment_method"` // BCA, OVO
	PaymentChannel string         `json:"payment_channel"`// VIRTUAL_ACCOUNT, QR_CODE
	PaymentStatus string          `json:"payment_status"` // PENDING, PAID
	
	// Transaction Details
	Amount        decimal.Decimal `json:"amount" gorm:"type:numeric(15,2)"`
	Currency      string          `json:"currency"`
	
	// Display Data (Hasil Initiate)
	PaymentCode   string          `json:"payment_code"`   // Nomor VA
	QrString      string          `json:"qr_string"`      // QR Raw String
	PaymentUrl    string          `json:"payment_url"`    // Redirect URL (E-wallet)
	ExpiryTime    *time.Time      `json:"expiry_time"`    // Kapan expire
	
	PaidAt        *time.Time      `json:"paid_at"`
	
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}