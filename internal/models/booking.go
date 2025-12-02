package models

import "time"

type TripType string
const (
	OneWay    TripType = "one_way"
	RoundTrip TripType = "round_trip"
)

type BookingStatus string
// Kamu bisa buat enum jika mau: pending, paid, cancelled

type Booking struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	UserID uint `json:"user_id"`
	User   User `json:"user"`

	BookingCode string `json:"booking_code" gorm:"uniqueIndex;not null"`
	BookingDate time.Time `json:"booking_date"`

	TripType        TripType `json:"trip_type" gorm:"type:trip_type"`
	TotalPassengers int      `json:"total_passengers"`
	TotalPrice      float64  `json:"total_price" gorm:"type:numeric(15,2)"`

	Status string `json:"status" gorm:"size:50;default:'pending'"`

	Details []BookingDetail `json:"details" gorm:"foreignKey:BookingID"`
}

func (Booking) TableName() string {
	return "bookings"
}
