package booking

import (
	"time"

	"github.com/shopspring/decimal" // [NEW] Import Library
)

type PassengerRequest struct {
	Title       string `json:"title" validate:"required,oneof=tuan nyonya nona mr ms mrs"`
	FullName    string `json:"full_name" validate:"required,min=2"`
	DOB         string `json:"dob" validate:"required,datetime=2006-01-02"` 
	Nationality string `json:"nationality" validate:"required"`
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

type BookingDetailResponse struct {
	BookingCode     string    `json:"booking_code"` 
	FlightCode      string    `json:"flight_code"`  
	Origin          string    `json:"origin"`
	Destination     string    `json:"destination"`
	DepartureTime   time.Time `json:"departure_time"`
	TotalPassengers int       `json:"total_passengers"`
	TotalPrice      decimal.Decimal `json:"total_price"`
}

type BookingResponse struct {
	OrderID         string    `json:"order_id"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	Status          string    `json:"status"`
	TransactionTime time.Time `json:"transaction_time"`
	PaymentURL      string    `json:"payment_url"` 
	ExpiryDate      *time.Time `json:"expiry_date"`
	Bookings        []BookingDetailResponse `json:"bookings"`
}