package booking

import (
	"errors"
	"fmt"
	"time"

	"ezytix-be/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var ErrBookingAlreadyCancelled = errors.New("booking already cancelled by scheduler")

type BookingRepository interface {
	CreateOrder(bookings []models.Booking) error
	GetBookingByOrderID(orderID string) (*models.Booking, error)
	FindBookingsByOrderID(orderID string) ([]models.Booking, error)
	UpdateBookingStatus(orderID string, status string) error
	UpdateBookingExpiry(orderID string, newExpiry time.Time) error
	GetExpiredBookings(currentTime time.Time) ([]models.Booking, error)
	CancelBookingAtomic(booking *models.Booking) error
	GetByUserID(userID uint) ([]models.Booking, error)
	UpdatePastBookingsToExpired() error
	GetBookingForInvoice(bookingCode string) (*models.Booking, error)
	GetPaymentByOrderID(orderID string) (*models.Payment, error)
	GetBookingForTicket(bookingCode string) (*models.Booking, error)
	GetBookingsForInvoiceByOrderID(orderID string) ([]models.Booking, error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db}
}

func (r *bookingRepository) GetBookingByOrderID(orderID string) (*models.Booking, error) {
	var booking models.Booking

	if err := r.db.Where("order_id = ?", orderID).First(&booking).Error; err != nil {
		return nil, err
	}

	var totalAmount float64
	err := r.db.Model(&models.Booking{}).
		Where("order_id = ?", orderID).
		Select("COALESCE(SUM(total_price), 0)").
		Scan(&totalAmount).Error

	if err != nil {
		return nil, err
	}

	booking.TotalPrice = decimal.NewFromFloat(totalAmount)

	return &booking, nil
}

func (r *bookingRepository) UpdateBookingExpiry(orderID string, newExpiry time.Time) error {
	return r.db.Model(&models.Booking{}).
		Where("order_id = ?", orderID).
		Update("expired_at", newExpiry).Error
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

func (r *bookingRepository) GetExpiredBookings(currentTime time.Time) ([]models.Booking, error) {
	var bookings []models.Booking
	
	err := r.db.Preload("Details").
		Where("status = ? AND expired_at < ?", models.BookingStatusPending, currentTime).
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
			Update("transaction_status", models.PaymentStatusExpire).Error; err != nil {
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

func (r *bookingRepository) GetByUserID(userID uint) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.
		Preload("Flight").
		Preload("Flight.Airline").
		Preload("Flight.OriginAirport").
		Preload("Flight.DestinationAirport").
		Preload("Flight.FlightClasses").
		Preload("Details").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bookings).Error

	if err != nil {
		return nil, err
	}
	return bookings, nil
}

func (r *bookingRepository) UpdatePastBookingsToExpired() error {
	query := `
		UPDATE bookings
		SET status = ?
		FROM flights
		WHERE bookings.flight_id = flights.id
		AND bookings.status = ?
		AND flights.arrival_time < ?
	`
	err := r.db.Exec(query, models.BookingStatusExpired, models.BookingStatusPaid, time.Now()).Error
	return err
}

func (r *bookingRepository) GetBookingForInvoice(bookingCode string) (*models.Booking, error) {
    var booking models.Booking

    err := r.db.
        Preload("User").                                      
        Preload("Details").                       
        Preload("Flight").
        Preload("Flight.FlightLegs").                         
        Preload("Flight.FlightLegs.Airline").               
        Preload("Flight.FlightLegs.OriginAirport").           
        Preload("Flight.FlightLegs.DestinationAirport").       
        Where("booking_code = ?", bookingCode).
        First(&booking).Error

    if err != nil {
        return nil, err
    }
    return &booking, nil
}

func (r *bookingRepository) GetPaymentByOrderID(orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *bookingRepository) GetBookingForTicket(bookingCode string) (*models.Booking, error) {
	var booking models.Booking

	err := r.db.
		Preload("User").
		Preload("Details").
		Preload("Flight").
		Preload("Flight.FlightLegs").
		Preload("Flight.FlightLegs.Airline").
		Preload("Flight.FlightLegs.OriginAirport").
		Preload("Flight.FlightLegs.DestinationAirport").
		Where("booking_code = ?", bookingCode).
		First(&booking).Error

	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) GetBookingsForInvoiceByOrderID(orderID string) ([]models.Booking, error) {
    var bookings []models.Booking

    err := r.db.
        Preload("User").
        Preload("Details").
        Preload("Flight").
        Preload("Flight.FlightLegs").
        Preload("Flight.FlightLegs.Airline").
        Preload("Flight.FlightLegs.OriginAirport").
        Preload("Flight.FlightLegs.DestinationAirport").
        Where("order_id = ?", orderID).
        Find(&bookings).Error

    if err != nil {
        return nil, err
    }
    return bookings, nil
}