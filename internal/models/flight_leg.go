package models

import "time"

// Flight Leg Model (Segmen Penerbangan)
type FlightLeg struct {
	ID                    uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FlightID              uint      `json:"flight_id"`
	LegOrder              int       `json:"leg_order"`
	DepartureTime         time.Time `json:"departure_time"`
	ArrivalTime           time.Time `json:"arrival_time"`
	OriginAirportID       uint      `json:"origin_airport_id"`
	DestinationAirportID  uint      `json:"destination_airport_id"`
	FlightNumber          string    `json:"flight_number"`
	AirlineName           string    `json:"airline_name"`
	AirlineLogo           string    `json:"airline_logo"`
	DepartureTerminal     string    `json:"departure_terminal"`
	ArrivalTerminal       string    `json:"arrival_terminal"`
	Duration              string    `json:"duration"`
	TransitNotes          string    `json:"transit_notes"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at"`
}

func (FlightLeg) TableName() string {
	return "flight_legs"
}
