package models

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	PaymentStatusPending    = "pending"
	PaymentStatusSettlement = "settlement"
	PaymentStatusExpire     = "expire"
	PaymentStatusDeny       = "deny"
	PaymentStatusCancel     = "cancel"
)

type Payment struct {
	ID                int             `json:"id" gorm:"primaryKey"`
	OrderID           string          `json:"order_id" gorm:"index"`
	TransactionID     string          `json:"transaction_id" gorm:"index"`
	PaymentType       string          `json:"payment_type"`              
	GrossAmount       decimal.Decimal `json:"gross_amount" gorm:"type:numeric(15,2)"`
	TransactionStatus string          `json:"transaction_status"`
	Bank     string `json:"bank"`
	VaNumber string `json:"va_number"`
	BillKey    string `json:"bill_key"`
	BillerCode string `json:"biller_code"`
	QrUrl    string `json:"qr_url" gorm:"type:text"`
	Deeplink string `json:"deeplink" gorm:"type:text"`
	ExpiryTime *time.Time `json:"expiry_time"`
	PaidAt     *time.Time `json:"paid_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}