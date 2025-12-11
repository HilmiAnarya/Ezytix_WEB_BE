package models

import "time"

// Flight Master Model (Penerbangan)
type Flight struct {
	ID                    uint         `json:"id" gorm:"primaryKey;autoIncrement"`
	FlightCode            string       `json:"flight_code" gorm:"unique;not null"`
	AirlineName           string       `json:"airline_name" gorm:"not null"`
	// Foreign Keys & Relations (PENTING UNTUK PRELOAD)
	OriginAirportID      uint    `json:"origin_airport_id"`
	OriginAirport        Airport `json:"origin_airport" gorm:"foreignKey:OriginAirportID"` // <--- TAMBAH INI

	DestinationAirportID uint    `json:"destination_airport_id"`
	DestinationAirport   Airport `json:"destination_airport" gorm:"foreignKey:DestinationAirportID"` // <--- TAMBAH INI
	
	DepartureTime         time.Time    `json:"departure_time" gorm:"not null"`
	ArrivalTime           time.Time    `json:"arrival_time" gorm:"not null"`
	TotalDuration         string       `json:"total_duration" gorm:"not null"`
	TransitCount          int          `json:"transit_count" gorm:"default:0"`
	TransitInfo           string       `json:"transit_info"`
	FlightLegs            []FlightLeg   `json:"flight_legs" gorm:"foreignKey:FlightID"`    // Relasi ke FlightLeg
	FlightClasses         []FlightClass `json:"flight_classes" gorm:"foreignKey:FlightID"` // Relasi ke FlightClass
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
	DeletedAt             *time.Time   `json:"deleted_at"`
}

func (Flight) TableName() string {
	return "flights"
}
