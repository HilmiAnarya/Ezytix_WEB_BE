package booking

import (
	"time"

	"github.com/shopspring/decimal"
)

// ==========================================
// REQUEST (INPUT) - TIDAK BERUBAH
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
// RESPONSE (OUTPUT) - UPDATED 🚀
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

type BookingResponse struct {
	OrderID         string                  `json:"order_id"`
	TotalAmount     decimal.Decimal         `json:"total_amount"`
	Status          string                  `json:"status"`
	TransactionTime time.Time               `json:"transaction_time"`
	ExpiryTime      *time.Time              `json:"expiry_time,omitempty"`
	Bookings        []BookingDetailResponse `json:"bookings"`
}

// [NEW] Struct untuk Detail Penumpang di History
type PassengerDetailResponse struct {
	FullName      string `json:"full_name"`
	Type          string `json:"type"`           // adult, child, infant
	TicketNumber  string `json:"ticket_number"`  // Nomor Tiket
	SeatClass     string `json:"seat_class"`
}

// [UPDATED] Response History
type MyBookingResponse struct {
	OrderID     string          `json:"order_id"`
	BookingCode string          `json:"booking_code"`
	Status      string          `json:"status"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiryTime  *time.Time      `json:"expiry_time,omitempty"`

	Flight     BookingFlightDetail       `json:"flight"`
	
	// [NEW] List Penumpang
	Passengers []PassengerDetailResponse `json:"passengers"` 
}

// [UPDATED] Detail Penerbangan
type BookingFlightDetail struct {
	FlightCode      string    `json:"flight_code"`
	AirlineName     string    `json:"airline_name"`
	AirlineLogo     string    `json:"airline_logo"`
	Origin          string    `json:"origin"`
	Destination     string    `json:"destination"`
	DepartureTime   time.Time `json:"departure_time"`
	ArrivalTime     time.Time `json:"arrival_time"`
	
	DurationMinutes int       `json:"duration_minutes"`
	
	// [NEW] Fields untuk mendukung UI baru
	DurationFormatted string  `json:"duration_formatted"` 
	TransitInfo       string  `json:"transit_info"`
	
	SeatClass       string    `json:"seat_class"`
	ClassCode       string    `json:"class_code"`
}