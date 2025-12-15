package payment

import (
	"ezytix-be/internal/models"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePayment(payment *models.Payment) error
	FindPaymentByOrderID(orderID string) (*models.Payment, error)
	FindPaymentByXenditID(xenditID string) (*models.Payment, error)
	UpdatePaymentStatus(orderID string, status string, paidAt *time.Time) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db}
}

// 1. Simpan Transaksi Baru (Pending)
func (r *paymentRepository) CreatePayment(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// 2. Cari Berdasarkan Order ID (Dipakai oleh Booking Service untuk cek status)
func (r *paymentRepository) FindPaymentByOrderID(orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// 3. Cari Berdasarkan Xendit ID (Dipakai saat Webhook masuk)
func (r *paymentRepository) FindPaymentByXenditID(xenditID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("xendit_id = ?", xenditID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// 4. Update Status (Jadi PAID, EXPIRED, atau FAILED)
func (r *paymentRepository) UpdatePaymentStatus(orderID string, status string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"payment_status": status,
		"updated_at":     time.Now(),
	}

	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	return r.db.Model(&models.Payment{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}