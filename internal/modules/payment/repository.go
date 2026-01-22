package payment

import (
	"ezytix-be/internal/models"
	"time"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	// Core CRUD
	CreatePayment(payment *models.Payment) error
	
	// Lookup Methods
	FindPaymentByOrderID(orderID string) (*models.Payment, error)
	FindPaymentByTransactionID(transactionID string) (*models.Payment, error) // [NEW] Penting untuk Webhook Midtrans
	
	// State Management
	UpdatePaymentStatus(orderID string, status string, paidAt *time.Time) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db}
}

// 1. Simpan Transaksi Baru
func (r *paymentRepository) CreatePayment(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// 2. Cari berdasarkan Order ID (Primary Lookup untuk User)
func (r *paymentRepository) FindPaymentByOrderID(orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// 3. Cari berdasarkan Transaction ID (Webhook Lookup)
// Webhook Midtrans sering mengirim Transaction ID sebagai identifier utama.
func (r *paymentRepository) FindPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("transaction_id = ?", transactionID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// 4. Update Status & Waktu Bayar
func (r *paymentRepository) UpdatePaymentStatus(orderID string, status string, paidAt *time.Time) error {
	updates := map[string]interface{}{
		"transaction_status": status, // Sesuaikan dengan nama kolom di DB/Model
		"updated_at":         time.Now(),
	}
	
	// Jika status settlement (lunas), catat waktu bayarnya
	if paidAt != nil {
		updates["paid_at"] = paidAt
	}

	return r.db.Model(&models.Payment{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}