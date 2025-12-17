package flight

import (
	"time"

	"github.com/shopspring/decimal" // [NEW] Import Library
)

// ==========================================
// 1. REQUEST DTO (Input dari User/Admin)
// ==========================================

type CreateFlightClassRequest struct {
	SeatClass  string          `json:"seat_class" validate:"required,oneof=economy business first_class"`
	Price      decimal.Decimal `json:"price" validate:"required"`
	TotalSeats int             `json:"total_seats" validate:"required,min=1"`
}

type CreateFlightLegRequest struct {
	LegOrder             int       `json:"leg_order" validate:"required,min=1"`
	OriginAirportID      uint      `json:"origin_airport_id" validate:"required"`
	DestinationAirportID uint      `json:"destination_airport_id" validate:"required"`
	DepartureTime        time.Time `json:"departure_time" validate:"required"`
	ArrivalTime          time.Time `json:"arrival_time" validate:"required"`
	FlightNumber         string    `json:"flight_number" validate:"required"`
	AirlineName          string    `json:"airline_name" validate:"required"`
	AirlineLogo          string    `json:"airline_logo"`
	DepartureTerminal    string    `json:"departure_terminal"`
	ArrivalTerminal      string    `json:"arrival_terminal"`
	Duration             string    `json:"duration" validate:"required"`
	TransitNotes         string    `json:"transit_notes"`
}

type CreateFlightRequest struct {
	FlightCode           string    `json:"flight_code" validate:"required"`
	AirlineName          string    `json:"airline_name" validate:"required"`
	OriginAirportID      uint      `json:"origin_airport_id" validate:"required"`
	DestinationAirportID uint      `json:"destination_airport_id" validate:"required"`
	DepartureTime        time.Time `json:"departure_time" validate:"required"`
	ArrivalTime          time.Time `json:"arrival_time" validate:"required"`
	TotalDuration        string    `json:"total_duration" validate:"required"`
	TransitInfo          string    `json:"transit_info"`

	Legs    []CreateFlightLegRequest   `json:"legs" validate:"required,min=1,dive"`
	Classes []CreateFlightClassRequest `json:"classes" validate:"required,min=1,dive"`
}

type SearchFlightRequest struct {
	OriginAirportID      uint   `query:"origin"`
	DestinationAirportID uint   `query:"destination"`
	DepartureDate        string `query:"date"`
	SeatClass            string `query:"seat_class"`
	PassengerCount       int    `query:"passengers"`
}

// ==========================================
// 2. RESPONSE DTO (Output ke Frontend)
// ==========================================
// Kita WAJIB tambahkan ini agar Service Refactor nanti tidak error.

type FlightClassResponse struct {
	SeatClass  string          `json:"seat_class"`
	Price      decimal.Decimal `json:"price"`
	TotalSeats int             `json:"total_seats"`
}

type FlightLegResponse struct {
	LegOrder             int       `json:"leg_order"`
	OriginAirportID      uint      `json:"origin_airport_id"`
	DestinationAirportID uint      `json:"destination_airport_id"`
	OriginCity           string    `json:"origin_city"`
	DestinationCity      string    `json:"destination_city"`
	DepartureTime        time.Time `json:"departure_time"`
	ArrivalTime          time.Time `json:"arrival_time"`
	FlightNumber         string    `json:"flight_number"`
	AirlineName          string    `json:"airline_name"`
	Duration             string    `json:"duration"`
}

type FlightResponse struct {
	ID                   uint                  `json:"id"`
	FlightCode           string                `json:"flight_code"`
	AirlineName          string                `json:"airline_name"`
	OriginCity           string                `json:"origin_city"`
	DestinationCity      string                `json:"destination_city"`
	DepartureTime        time.Time             `json:"departure_time"`
	ArrivalTime          time.Time             `json:"arrival_time"`
	TotalDuration        string                `json:"total_duration"`
	TransitInfo          string                `json:"transit_info"`
	Classes              []FlightClassResponse `json:"classes"`
	Legs                 []FlightLegResponse   `json:"legs"`
}