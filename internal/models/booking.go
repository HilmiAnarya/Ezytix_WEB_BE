package models

import "time"

const (
	BookingStatusPending   = "pending"
	BookingStatusPaid      = "paid"
	BookingStatusCancelled = "cancelled"
	BookingStatusFailed    = "failed"

	TripTypeOneWay    = "one_way"
	TripTypeRoundTrip = "round_trip"
)

type Booking struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	// [THE GLUE] Grouping ID (Satu Order ID bisa punya banyak Booking ID untuk RoundTrip)
	OrderID string `json:"order_id" gorm:"size:50;not null;index"`

	// Foreign Keys
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user" gorm:"foreignKey:UserID"`

	FlightID uint   `json:"flight_id" gorm:"not null"`
	Flight   Flight `json:"flight" gorm:"foreignKey:FlightID"`

	// Transaction Info
	BookingCode string `json:"booking_code" gorm:"size:20;uniqueIndex;not null"`
	
	// Metadata Trip
	TripType string `json:"trip_type" gorm:"type:trip_type;default:'one_way';not null"`

	// Snapshot Data
	TotalPassengers int     `json:"total_passengers" gorm:"not null"`
	TotalPrice      float64 `json:"total_price" gorm:"type:numeric(15,2);not null"`
	Status          string  `json:"status" gorm:"size:20;default:'pending';not null"`

	// Relation to Details (Child)
	Details []BookingDetail `json:"details" gorm:"foreignKey:BookingID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Booking) TableName() string {
	return "bookings"
}