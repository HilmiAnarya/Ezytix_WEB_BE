package flight

import "time"

// ==============================
// REQUEST DTO
// ==============================

// 1. DTO CHILD: FlightLeg (Menyesuaikan models/flight_leg.go)
type CreateFlightLegRequest struct {
	LegOrder             int       `json:"leg_order" validate:"required,min=1"`
	OriginAirportID      uint      `json:"origin_airport_id" validate:"required"`
	DestinationAirportID uint      `json:"destination_airport_id" validate:"required"`
	DepartureTime        time.Time `json:"departure_time" validate:"required"`
	ArrivalTime          time.Time `json:"arrival_time" validate:"required"`
	
	FlightNumber         string    `json:"flight_number" validate:"required"` // e.g. "GA-404"
	AirlineName          string    `json:"airline_name" validate:"required"`
	AirlineLogo          string    `json:"airline_logo"`       // Opsional (URL gambar)
	
	DepartureTerminal    string    `json:"departure_terminal"` // Opsional, e.g. "Terminal 3"
	ArrivalTerminal      string    `json:"arrival_terminal"`   // Opsional
	
	Duration             string    `json:"duration" validate:"required"` // e.g. "1j 45m"
	TransitNotes         string    `json:"transit_notes"`      // Opsional
}

// 2. DTO CHILD: FlightClass (Menyesuaikan models/flight_class.go)
type CreateFlightClassRequest struct {
	// Enum validation penting agar tidak sembarang string masuk
	SeatClass string  `json:"seat_class" validate:"required,oneof=economy business first_class"` 
	Price     float64 `json:"price" validate:"required,min=0"`
}

// 3. DTO PARENT: Flight (Menyesuaikan models/flight.go)
type CreateFlightRequest struct {
	FlightCode           string    `json:"flight_code" validate:"required"` // Kode unik
	AirlineName          string    `json:"airline_name" validate:"required"`
	
	OriginAirportID      uint      `json:"origin_airport_id" validate:"required"`
	DestinationAirportID uint      `json:"destination_airport_id" validate:"required"`
	
	DepartureTime        time.Time `json:"departure_time" validate:"required"`
	ArrivalTime          time.Time `json:"arrival_time" validate:"required"`
	
	TotalDuration        string    `json:"total_duration" validate:"required"`
	//TransitCount         int       `json:"transit_count" validate:"min=0"`
	TransitInfo          string    `json:"transit_info"` // Opsional, e.g. "1 Stop"
	
	// Validasi Nested Structure (Dive)
	Legs    []CreateFlightLegRequest   `json:"legs" validate:"required,min=1,dive"` 
	Classes []CreateFlightClassRequest `json:"classes" validate:"required,min=1,dive"`
}