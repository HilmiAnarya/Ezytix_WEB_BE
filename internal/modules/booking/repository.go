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
	
	// [NEW] Untuk kebutuhan Payment Service (Single Object)
	GetBookingByOrderID(orderID string) (*models.Booking, error)
	
	// [LEGACY] Untuk kebutuhan internal booking (List)
	FindBookingsByOrderID(orderID string) ([]models.Booking, error)
	
	UpdateBookingStatus(orderID string, status string) error
	
	// [NEW] Update ExpiredAt saat initiate payment
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

	// 1. Ambil Data Booking Pertama (untuk info UserID, ExpiredAt, FlightID, dll sebagai referensi)
	if err := r.db.Where("order_id = ?", orderID).First(&booking).Error; err != nil {
		return nil, err
	}

	// 2. [CRITICAL FIX] Hitung Total Harga dari SEMUA booking dengan OrderID ini
	// Skenario: Round Trip (Pergi + Pulang) -> Ada 2 row di database.
	// Kita harus menjumlahkan total_price keduanya agar User membayar full.
	var totalAmount float64
	err := r.db.Model(&models.Booking{}).
		Where("order_id = ?", orderID).
		Select("COALESCE(SUM(total_price), 0)"). // COALESCE agar tidak error jika null
		Scan(&totalAmount).Error

	if err != nil {
		return nil, err
	}

	// 3. Override TotalPrice di struct booking dengan hasil penjumlahan
	// Convert float64 ke decimal agar tipe datanya sesuai struct Payment Service
	booking.TotalPrice = decimal.NewFromFloat(totalAmount)

	return &booking, nil
}

// [NEW IMPLEMENTATION] Update Expiry
func (r *bookingRepository) UpdateBookingExpiry(orderID string, newExpiry time.Time) error {
	// Update ExpiredAt menjadi waktu absolut dari Xendit
	return r.db.Model(&models.Booking{}).
		Where("order_id = ?", orderID).
		Update("expired_at", newExpiry).Error
}

// --- EXISTING FUNCTIONS (REFACTORED FOR STRICT EXPIRY) ---

func (r *bookingRepository) CreateOrder(bookings []models.Booking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range bookings {
			booking := &bookings[i]

			if len(booking.Details) == 0 {
				return errors.New("booking details/passengers cannot be empty")
			}

			seatClass := booking.Details[0].SeatClass
			passengerCount := len(booking.Details)

			// Decrement Flight Class Seats
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

// [REFACTORED] Menggunakan expired_at bukan created_at
func (r *bookingRepository) GetExpiredBookings(currentTime time.Time) ([]models.Booking, error) {
	var bookings []models.Booking
	
	// Query: Cari booking PENDING yang expired_at < NOW()
	// Jika expired_at NULL (migrasi lama), fallback ke created_at logic (opsional, tapi sebaiknya strict)
	err := r.db.Preload("Details").
		Where("status = ? AND expired_at < ?", models.BookingStatusPending, currentTime).
		Find(&bookings).Error
		
	return bookings, err
}

func (r *bookingRepository) CancelBookingAtomic(booking *models.Booking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Update Booking Status
		if err := tx.Model(booking).Update("status", models.BookingStatusCancelled).Error; err != nil {
			return err
		}

		// 2. Update Payment Status (Jika ada) - Menggunakan OrderID
		if err := tx.Model(&models.Payment{}).
			Where("order_id = ?", booking.OrderID).
			Update("transaction_status", models.PaymentStatusExpire).Error; err != nil {
			// Ignore error if payment not found, just continue cancellation
		}

		// 3. Restock Seats
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

    // Eager Loading Gila-gilaan untuk kebutuhan PDF
    err := r.db.
        Preload("User").                                       // Data Pemesan
        Preload("Details").                             // Data Penumpang & Harga per item
        Preload("Flight").                                     // Data Penerbangan Utama
        //Preload("Flight.FlightClass").                         // Kelas (Economy, etc)
        Preload("Flight.FlightLegs").                          // Segmen Penerbangan
        Preload("Flight.FlightLegs.Airline").                  // Nama Maskapai
        Preload("Flight.FlightLegs.OriginAirport").            // Bandara Asal
        Preload("Flight.FlightLegs.DestinationAirport").       // Bandara Tujuan
        Where("booking_code = ?", bookingCode).
        First(&booking).Error

    if err != nil {
        return nil, err
    }
    return &booking, nil
}

func (r *bookingRepository) GetPaymentByOrderID(orderID string) (*models.Payment, error) {
	var payment models.Payment
	// Cari berdasarkan kolom order_id
	err := r.db.Where("order_id = ?", orderID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil jika belum ada payment (aman)
		}
		return nil, err
	}
	return &payment, nil
}

func (r *bookingRepository) GetBookingForTicket(bookingCode string) (*models.Booking, error) {
	var booking models.Booking

	// Kita butuh data Flight, Legs, Airline, Airport, User, dan Details
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

    // Gunakan Find (bukan First) karena hasilnya bisa lebih dari 1 (misal: Round Trip)
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