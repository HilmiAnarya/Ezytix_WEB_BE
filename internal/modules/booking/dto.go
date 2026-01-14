package booking

import (
	"time"

	"github.com/shopspring/decimal"
)

// ==========================================
// REQUEST (INPUT)
// ==========================================

type PassengerRequest struct {
	Title          string `json:"title" validate:"required,oneof=tuan nyonya nona mr ms mrs"`
	FullName       string `json:"full_name" validate:"required,min=2"`
	DOB            string `json:"dob" validate:"required,datetime=2006-01-02"`
	Nationality    string `json:"nationality" validate:"required"`
	PassportNumber string `json:"passport_number"`
	IssuingCountry string `json:"issuing_country"`
	ValidUntil     string `json:"valid_until" validate:"omitempty,datetime=2006-01-02"`
}

type BookingItemRequest struct {
	FlightID   uint               `json:"flight_id" validate:"required"`
	SeatClass  string             `json:"seat_class" validate:"required,oneof=economy business first_class"`
	Passengers []PassengerRequest `json:"passengers" validate:"required,min=1,dive"`
}

type CreateOrderRequest struct {
	Items []BookingItemRequest `json:"items" validate:"required,min=1,dive"`
}

// ==========================================
// RESPONSE (OUTPUT)
// ==========================================

type BookingDetailResponse struct {
	BookingCode     string          `json:"booking_code"`
	FlightCode      string          `json:"flight_code"`
	Origin          string          `json:"origin"`
	Destination     string          `json:"destination"`
	DepartureTime   time.Time       `json:"departure_time"`
	TotalPassengers int             `json:"total_passengers"`
	TotalPrice      decimal.Decimal `json:"total_price"`
}

// [REFACTORED] Response Create Booking
// Hapus PaymentURL karena pembayaran dilakukan di step terpisah
type BookingResponse struct {
	OrderID         string                  `json:"order_id"` // Kunci utama untuk redirect ke payment page
	TotalAmount     decimal.Decimal         `json:"total_amount"`
	Status          string                  `json:"status"` // Akan selalu "pending" saat baru dibuat
	TransactionTime time.Time               `json:"transaction_time"`
	ExpiryTime 		*time.Time `json:"expiry_time,omitempty"` // Batas waktu pembayaran (sebelum scheduler jalan)
	Bookings        []BookingDetailResponse `json:"bookings"`
}

// [REFACTORED] Response History
type MyBookingResponse struct {
	OrderID     string          `json:"order_id"` // [ADDED] Penting untuk tombol "Bayar" di history
	BookingCode string          `json:"booking_code"`
	Status      string          `json:"status"` // pending, paid, cancelled, failed
	TotalAmount decimal.Decimal `json:"total_amount"`
	CreatedAt   time.Time       `json:"created_at"`

	// [REMOVED] PaymentUrl dihapus.
	// Jika status "pending", Frontend akan pakai BookingCode untuk redirect ke halaman /payment/:id/select
	ExpiryTime *time.Time `json:"expiry_time,omitempty"`

	// Data Penerbangan
	Flight BookingFlightDetail `json:"flight"`
}

type BookingFlightDetail struct {
	FlightCode      string    `json:"flight_code"`
	AirlineName     string    `json:"airline_name"`
	AirlineLogo     string    `json:"airline_logo"`
	Origin          string    `json:"origin"`
	Destination     string    `json:"destination"`
	DepartureTime   time.Time `json:"departure_time"`
	ArrivalTime     time.Time `json:"arrival_time"`
	DurationMinutes int       `json:"duration_minutes"`
	SeatClass       string    `json:"seat_class"`
	ClassCode       string    `json:"class_code"`
}