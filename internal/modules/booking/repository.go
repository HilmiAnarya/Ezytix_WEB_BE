package booking

import (
	"errors"
	"fmt"
	"ezytix-be/internal/models"

	"gorm.io/gorm"
)

type BookingRepository interface {
	// Transactional: Kurangi Stok -> Simpan Booking -> Simpan Details
	CreateOrder(bookings []models.Booking) error

	// Helper untuk Webhook & History
	FindBookingsByOrderID(orderID string) ([]models.Booking, error)
	UpdateBookingStatus(orderID string, status string) error
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db}
}

// ==========================================
// 1. ATOMIC CREATE ORDER
// ==========================================
func (r *bookingRepository) CreateOrder(bookings []models.Booking) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Kita loop setiap booking (Flight) dalam order ini
		// (Ingat: Round Trip = 2 Booking Objects)
		for i := range bookings {
			booking := &bookings[i] // Pakai pointer ke elemen slice agar efisien

			// A. Safety Check
			if len(booking.Details) == 0 {
				return errors.New("booking details/passengers cannot be empty")
			}

			// Ambil info kelas & jumlah dari penumpang pertama (karena 1 booking = 1 kelas)
			seatClass := booking.Details[0].SeatClass
			passengerCount := len(booking.Details)

			// B. ATOMIC STOCK DEDUCTION (The "Lock")
			// Query: UPDATE flight_classes SET total_seats = total_seats - N
			//        WHERE flight_id = ? AND seat_class = ? AND total_seats >= N
			// Logic: Ini mencegah Race Condition. Jika stok cukup, langsung kurangi.
			result := tx.Model(&models.FlightClass{}).
				Where("flight_id = ? AND seat_class = ? AND total_seats >= ?",
					booking.FlightID, seatClass, passengerCount).
				Update("total_seats", gorm.Expr("total_seats - ?", passengerCount))

			if result.Error != nil {
				return result.Error // Error DB teknis
			}

			// Jika tidak ada baris yang terupdate, berarti STOK TIDAK CUKUP
			if result.RowsAffected == 0 {
				return fmt.Errorf("insufficient stock for flight ID %d class %s", booking.FlightID, seatClass)
			}

			// C. INSERT BOOKING (Header & Details)
			// GORM cukup pintar: Saat kita Create Parent (Booking),
			// dia otomatis Create Children (Details) karena ada di struct.
			if err := tx.Create(booking).Error; err != nil {
				return err
			}
		}

		return nil // Commit Transaction (Semua sukses)
	})
}

// ==========================================
// 2. FIND BY ORDER ID
// ==========================================
func (r *bookingRepository) FindBookingsByOrderID(orderID string) ([]models.Booking, error) {
	var bookings []models.Booking
	// Preload flight dan details untuk keperluan invoice/email nanti
	err := r.db.Preload("Details").Preload("Flight").
		Where("order_id = ?", orderID).
		Find(&bookings).Error

	return bookings, err
}

// ==========================================
// 3. UPDATE STATUS (Bulk Update)
// ==========================================
func (r *bookingRepository) UpdateBookingStatus(orderID string, status string) error {
	// Update semua booking yang punya Order ID sama (Pergi & Pulang)
	return r.db.Model(&models.Booking{}).
		Where("order_id = ?", orderID).
		Update("status", status).Error
}