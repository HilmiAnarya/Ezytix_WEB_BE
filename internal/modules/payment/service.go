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

type BookingServiceContract interface {
	GetBookingByOrderID(orderID string) (*models.Booking, error)
	UpdateBookingStatus(orderID string, status string) error
}

type PaymentService interface {
	InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error)
	ProcessWebhook(payload map[string]interface{}) error
	CancelPayment(orderID string) error
	GetPaymentByOrderID(orderID string) (*InitiatePaymentResponse, error)
}

type paymentService struct {
	repo           PaymentRepository
	bookingRepo    BookingServiceContract
	midtransClient coreapi.Client
}

func NewPaymentService(repo PaymentRepository, bookingRepo BookingServiceContract) PaymentService {

	var client coreapi.Client
	
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

func (s *paymentService) InitiatePayment(req InitiatePaymentRequest) (*InitiatePaymentResponse, error) {

	booking, err := s.bookingRepo.GetBookingByOrderID(req.OrderID)
	if err != nil {
		return nil, errors.New("booking not found")
	}
	if booking.Status == models.BookingStatusPaid {
		return nil, errors.New("booking already paid")
	}

	if booking.ExpiredAt == nil {
		return nil, errors.New("booking expiry data is invalid")
	}

	if time.Now().After(*booking.ExpiredAt) {
		return nil, errors.New("booking expired")
	}

	existing, _ := s.repo.FindPaymentByOrderID(req.OrderID)
	if existing != nil && existing.TransactionStatus == "pending" {
		if existing.PaymentType == req.PaymentType {
			return s.constructResponseFromModel(existing), nil
		}
	}

	remainingDuration := booking.ExpiredAt.Sub(time.Now())
	
	minutesLeft := int(remainingDuration.Minutes())

	if minutesLeft < 1 {
		return nil, errors.New("booking time is almost up, please re-book")
	}

	now := time.Now()
	midtransTimeFormat := now.Format("2006-01-02 15:04:05 -0700")

	grossAmt := int64(booking.TotalPrice.InexactFloat64())
	
	chargeReq := &coreapi.ChargeReq{
		PaymentType: coreapi.CoreapiPaymentType(req.PaymentType),
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  req.OrderID,
			GrossAmt: grossAmt,
		},
		CustomExpiry: &coreapi.CustomExpiry{
			OrderTime:      midtransTimeFormat,
			ExpiryDuration: minutesLeft,        
			Unit:           "minute",
		},
		Metadata: map[string]interface{}{
			"user_id": fmt.Sprintf("%d", booking.UserID),
		},
	}

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

	resp, midtransErr := s.midtransClient.ChargeTransaction(chargeReq)
	if midtransErr != nil {
		return nil, fmt.Errorf("midtrans error: %v", midtransErr)
	}

	return s.saveAndRespond(booking, req, resp, booking.ExpiredAt)
}

func (s *paymentService) saveAndRespond(booking *models.Booking, req InitiatePaymentRequest, resp *coreapi.ChargeResponse, strictExpiry *time.Time) (*InitiatePaymentResponse, error) {
	var vaNumber, bank, billKey, billerCode, qrUrl, deeplink string

	switch req.PaymentType {
	case "bank_transfer":
		if len(resp.VaNumbers) > 0 {
			vaNumber = resp.VaNumbers[0].VANumber
			bank = resp.VaNumbers[0].Bank
		} else if resp.PermataVaNumber != "" {
			vaNumber = resp.PermataVaNumber
			bank = "permata"
		}
	case "echannel":
		billKey = resp.BillKey
		billerCode = resp.BillerCode
	case "qris", "gopay":
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
		GrossAmount:       booking.TotalPrice,
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

	return s.constructResponseFromModel(paymentModel), nil
}

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

func (s *paymentService) ProcessWebhook(payload map[string]interface{}) error {
	orderID, _ := payload["order_id"].(string)
	transactionID, _ := payload["transaction_id"].(string)
	statusCode, _ := payload["status_code"].(string)
	grossAmount, _ := payload["gross_amount"].(string)
	signatureKey, _ := payload["signature_key"].(string)
	transactionStatus, _ := payload["transaction_status"].(string)

	input := orderID + statusCode + grossAmount + config.AppConfig.MidtransServerKey
	hash := sha512.Sum512([]byte(input))
	if signatureKey != hex.EncodeToString(hash[:]) {
		return errors.New("invalid signature key")
	}

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

	var paidAt *time.Time
	if isPaid {
		now := time.Now()
		paidAt = &now
	}

	if err := s.repo.UpdatePaymentStatusByTransactionID(transactionID, internalStatus, paidAt); err != nil {
		return err
	}

	if isPaid {
		return s.bookingRepo.UpdateBookingStatus(orderID, models.BookingStatusPaid)
	}

	return nil
}

func (s *paymentService) CancelPayment(orderID string) error {
	payment, err := s.repo.FindPaymentByOrderID(orderID)
	if err != nil {
		return err
	}

	isPending := payment.TransactionStatus == models.PaymentStatusPending || payment.TransactionStatus == "waiting"

	if isPending {
		fmt.Printf("🚀 Cancelling Transaction ID: %s for Order: %s\n", payment.TransactionID, orderID)
		
		_, midErr := s.midtransClient.CancelTransaction(payment.TransactionID)
		if midErr != nil {
			fmt.Printf("⚠️ Midtrans Cancel Note: %v\n", midErr)
		}

		return s.repo.UpdatePaymentStatusByTransactionID(payment.TransactionID, models.PaymentStatusCancel, nil)
	}

	return nil
}

func (s *paymentService) GetPaymentByOrderID(orderID string) (*InitiatePaymentResponse, error) {
	payment, err := s.repo.FindPaymentByOrderID(orderID)
	if err != nil {
		return nil, err
	}
	resp := &InitiatePaymentResponse{
		OrderID:           payment.OrderID,
		TransactionID:     payment.TransactionID,
		PaymentType:       payment.PaymentType,
		Amount:            payment.GrossAmount.InexactFloat64(), 
		TransactionStatus: payment.TransactionStatus,
	}

	if payment.ExpiryTime != nil {
		resp.ExpiryTime = *payment.ExpiryTime
	}

	switch payment.PaymentType {
	case "bank_transfer":
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