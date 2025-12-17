  package models

import "time"

type FlightLeg struct {
	ID                    uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FlightID              uint      `json:"flight_id"`
	
	LegOrder              int       `json:"leg_order"`

	AirlineID 			uint     `gorm:"not null;index" json:"airline_id"` // Operating Carrier
	Airline   			*Airline `gorm:"foreignKey:AirlineID" json:"airline,omitempty"`

	DepartureTime         time.Time `json:"departure_time"`
	ArrivalTime           time.Time `json:"arrival_time"`

	OriginAirportID      uint      `gorm:"not null" json:"origin_airport_id"`
	OriginAirport        *Airport `gorm:"foreignKey:OriginAirportID" json:"origin_airport,omitempty"`
	DestinationAirportID uint      `gorm:"not null" json:"destination_airport_id"`
	DestinationAirport   *Airport `gorm:"foreignKey:DestinationAirportID" json:"destination_airport,omitempty"`

	FlightNumber          string    `json:"flight_number"`
	
	Duration     		 int       `json:"duration"` // Dalam menit
	TransitNotes          string    `json:"transit_notes"`

	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	DeletedAt             *time.Time `json:"deleted_at"`
}

func (FlightLeg) TableName() string {
	return "flight_legs"
}
