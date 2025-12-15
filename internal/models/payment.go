package models

import "time"

const (
	PaymentStatusPending = "PENDING"
	PaymentStatusPaid    = "PAID"
	PaymentStatusExpired = "EXPIRED"
	PaymentStatusFailed  = "FAILED"
)

type Payment struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderID string `json:"order_id" gorm:"size:50;not null;index"`
	XenditID      string `json:"xendit_id" gorm:"size:100;index"`
	PaymentMethod string `json:"payment_method" gorm:"size:50;default:'QRIS'"`
	PaymentStatus string `json:"payment_status" gorm:"size:20;default:'PENDING';not null"`
	Amount   float64 `json:"amount" gorm:"type:numeric(15,2);not null"`
	Currency string  `json:"currency" gorm:"size:3;default:'IDR'"`
	PaymentURL string `json:"payment_url" gorm:"type:text"`
	PaidAt    *time.Time `json:"paid_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}