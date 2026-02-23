package payment

import "time"

type InitiatePaymentRequest struct {
	OrderID     string `json:"order_id" validate:"required"`
	PaymentType string `json:"payment_type" validate:"required,oneof=bank_transfer echannel qris gopay"` 
	Bank        string `json:"bank" validate:"required_if=PaymentType bank_transfer"` 
}

type InitiatePaymentResponse struct {
	OrderID           string    `json:"order_id"`
	TransactionID     string    `json:"transaction_id"`
	PaymentType       string    `json:"payment_type"`
	Amount            float64   `json:"amount"`
	TransactionStatus string    `json:"transaction_status"`
	ExpiryTime        time.Time `json:"expiry_time"`
	VirtualAccount *VirtualAccountData `json:"virtual_account,omitempty"`
	MandiriBill    *MandiriBillData    `json:"mandiri_bill,omitempty"`
	Qris           *QrisData           `json:"qris,omitempty"`
	Gopay          *GopayData          `json:"gopay,omitempty"`
}

type VirtualAccountData struct {
	Bank     string `json:"bank"`
	VaNumber string `json:"va_number"`
}

type MandiriBillData struct {
	BillKey    string `json:"bill_key"`
	BillerCode string `json:"biller_code"`
}

type QrisData struct {
	QrUrl string `json:"qr_url"`
}

type GopayData struct {
	Deeplink string `json:"deeplink"`
}