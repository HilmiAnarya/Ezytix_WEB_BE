package booking

import (
	"errors"
	"fmt"
	"ezytix-be/internal/models"
	"time" 

	"gorm.io/gorm"
)

var ErrBookingAlreadyCancelled = errors.New("booking already cancelled by scheduler")
type BookingRepository interface {
	CreateOrder(bookings []models.Booking) error
	FindBookingsByOrderID(orderID string) ([]models.Booking, error)
	UpdateBookingStatus(orderID string, status string) error
	GetExpiredBookings(expiryTime time.Time) ([]models.Booking, error)
	CancelBookingAtomic(booking *models.Booking) error
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db}
}

func (r *bookingRepository) CreateOrder(bookings []models.Booking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range bookings {
			booking := &bookings[i] 

			if len(booking.Details) == 0 {
				return errors.New("booking details/passengers cannot be empty")
			}

			seatClass := booking.Details[0].SeatClass
			passengerCount := len(booking.Details)

			result := tx.Model(&models.FlightClass{}).
				Where("flight_id = ? AND seat_class = ? AND total_seats >= ?",
					booking.FlightID, seatClass, passengerCount).
				Update("total_seats", gorm.Expr("total_seats - ?", passengerCount))

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				return fmt.Errorf("insufficient stock for flight ID %d class %s", booking.FlightID, seatClass)
			}

			if err := tx.Create(booking).Error; err != nil {
				return err
			}
		}

		return nil 
	})
}

func (r *bookingRepository) FindBookingsByOrderID(orderID string) ([]models.Booking, error) {
	var bookings []models.Booking

	err := r.db.Preload("Details").Preload("Flight").
		Where("order_id = ?", orderID).
		Find(&bookings).Error

	return bookings, err
}

func (r *bookingRepository) UpdateBookingStatus(orderID string, status string) error {
	var currentStatus string
	
	err := r.db.Model(&models.Booking{}).
		Select("status").
		Where("order_id = ?", orderID).
		Scan(&currentStatus).Error

	if err != nil {
		return err 
	}

	
	if currentStatus == models.BookingStatusCancelled {
		return ErrBookingAlreadyCancelled
	}

	return r.db.Model(&models.Booking{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error
}

func (r *bookingRepository) GetExpiredBookings(expiryTime time.Time) ([]models.Booking, error) {
	var bookings []models.Booking

	err := r.db.Preload("Details").
		Where("status = ? AND created_at < ?", models.BookingStatusPending, expiryTime).
		Find(&bookings).Error
		
	return bookings, err
}

func (r *bookingRepository) CancelBookingAtomic(booking *models.Booking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(booking).Update("status", models.BookingStatusCancelled).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Payment{}).
			Where("order_id = ?", booking.OrderID).
			Update("payment_status", models.PaymentStatusExpired).Error; err != nil {
			return err
		}

		if len(booking.Details) > 0 {
			seatClass := booking.Details[0].SeatClass
			passengerCount := len(booking.Details)

			if err := tx.Model(&models.FlightClass{}).
				Where("flight_id = ? AND seat_class = ?", booking.FlightID, seatClass).
				Update("total_seats", gorm.Expr("total_seats + ?", passengerCount)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}