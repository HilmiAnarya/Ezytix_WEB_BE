package admin

import (
	"ezytix-be/internal/models"
	"time"

	"gorm.io/gorm"
)

type AdminRepository interface {
	CountCustomers() (int64, error)
	CountBookingsToday() (int64, error)
	SumRevenueToday() (float64, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db}
}

func (r *adminRepository) CountCustomers() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("role = ?", models.RoleCustomer).Count(&count).Error
	return count, err
}

func (r *adminRepository) CountBookingsToday() (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&models.Booking{}).Where("created_at >= ?", today).Count(&count).Error
	return count, err
}

func (r *adminRepository) SumRevenueToday() (float64, error) {
	var total float64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&models.Booking{}).
		Where("created_at >= ? AND status = ?", today, models.BookingStatusPaid).
		Select("COALESCE(SUM(total_price), 0)").Scan(&total).Error
	return total, err
}