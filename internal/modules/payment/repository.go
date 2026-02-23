package payment

import (
	"ezytix-be/internal/models"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePayment(payment *models.Payment) error
	FindPaymentByOrderID(orderID string) (*models.Payment, error)
	FindPaymentByTransactionID(transactionID string) (*models.Payment, error)
	UpdatePaymentStatus(orderID string, status string, paidAt *time.Time) error
	UpdatePaymentStatusByTransactionID(transactionID string, status string, paidAt *time.Time) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db}
}

func (r *paymentRepository) CreatePayment(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) FindPaymentByOrderID(orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("order_id = ?", orderID).Order("id desc").First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("transaction_id = ?", transactionID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}


func (r *paymentRepository) UpdatePaymentStatus(orderID string, status string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"transaction_status": status,
		"updated_at":         time.Now(),
	}

	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	return r.db.Model(&models.Payment{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *paymentRepository) UpdatePaymentStatusByTransactionID(transactionID string, status string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"transaction_status": status,
		"updated_at":         time.Now(),
	}

	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	return r.db.Model(&models.Payment{}).Where("transaction_id = ?", transactionID).Updates(updates).Error
}