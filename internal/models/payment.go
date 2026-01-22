package models

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	// Mapping Status Midtrans ke Internal App
	PaymentStatusPending    = "pending"
	PaymentStatusSettlement = "settlement" // Paid
	PaymentStatusExpire     = "expire"
	PaymentStatusDeny       = "deny"
	PaymentStatusCancel     = "cancel"
)

type Payment struct {
	ID                int             `json:"id" gorm:"primaryKey"`
	OrderID           string          `json:"order_id" gorm:"index"`
	
	// [MIDTRANS CORE]
	TransactionID     string          `json:"transaction_id" gorm:"index"` // UUID Midtrans
	PaymentType       string          `json:"payment_type"`                // bank_transfer, echannel, qris, gopay
	GrossAmount       decimal.Decimal `json:"gross_amount" gorm:"type:numeric(15,2)"`
	TransactionStatus string          `json:"transaction_status"`          // Status dari Midtrans

	// [BANK TRANSFER] (BCA, BNI, BRI, Permata)
	Bank     string `json:"bank"`
	VaNumber string `json:"va_number"`

	// [MANDIRI BILL]
	BillKey    string `json:"bill_key"`
	BillerCode string `json:"biller_code"`

	// [QRIS & E-WALLET]
	QrUrl    string `json:"qr_url" gorm:"type:text"`   // URL Gambar QR
	Deeplink string `json:"deeplink" gorm:"type:text"` // Link Redirect App

	// [TIMESTAMPS]
	ExpiryTime *time.Time `json:"expiry_time"`
	PaidAt     *time.Time `json:"paid_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}