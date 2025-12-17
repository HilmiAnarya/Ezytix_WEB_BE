package models

import "time"

type Flight struct {
	ID                    uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	FlightCode            string       `json:"flight_code" gorm:"unique;not null"`

	AirlineID    			uint      `gorm:"not null;index" json:"airline_id"` // FK ke tabel airlines
	Airline      			*Airline  `gorm:"foreignKey:AirlineID" json:"airline,omitempty"` // Relasi untuk Preload

	OriginAirportID       uint    		`json:"origin_airport_id"`
	OriginAirport         *Airport 		`json:"origin_airport" gorm:"foreignKey:OriginAirportID"`
	DestinationAirportID  uint    		`json:"destination_airport_id"`
	DestinationAirport    *Airport 		`json:"destination_airport" gorm:"foreignKey:DestinationAirportID"`

	DepartureTime         time.Time    `json:"departure_time" gorm:"not null"`
	ArrivalTime           time.Time    `json:"arrival_time" gorm:"not null"`

	TotalDuration int `gorm:"not null" json:"total_duration"` // Dalam menit

	TransitCount          int          `json:"transit_count" gorm:"default:0"`
	TransitInfo           string       `json:"transit_info"`

	FlightLegs            []FlightLeg   `json:"flight_legs" gorm:"foreignKey:FlightID"`
	FlightClasses         []FlightClass `json:"flight_classes" gorm:"foreignKey:FlightID"`

	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
	DeletedAt             *time.Time   `json:"deleted_at"`
}

func (Flight) TableName() string {
	return "flights"
}
