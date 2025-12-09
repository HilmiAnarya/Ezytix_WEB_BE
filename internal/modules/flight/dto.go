package flight

import "time"

// ==============================
// 1. REQUEST DTO (ADMIN INPUT)
// ==============================

type CreateFlightClassRequest struct {
	SeatClass  string  `json:"seat_class" validate:"required,oneof=economy business first_class"`
	Price      float64 `json:"price" validate:"required,min=0"`
	// [NEW] Admin wajib isi kuota kursi
	TotalSeats int     `json:"total_seats" validate:"required,min=1"` 
}

// ... CreateFlightLegRequest TETAP SAMA seperti sebelumnya ...
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

// ==============================
// 2. SEARCH DTO (USER FILTER)
// ==============================

// Mencakup semua filter dari UI
type SearchFlightRequest struct {
	OriginAirportID      uint   `query:"origin"`
	DestinationAirportID uint   `query:"destination"`
	DepartureDate        string `query:"date"`       // YYYY-MM-DD
	SeatClass            string `query:"seat_class"` // economy, business
	PassengerCount       int    `query:"passengers"` // min 1
}