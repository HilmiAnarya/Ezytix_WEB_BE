package models

import "time"

const (
	PassengerTypeAdult  = "adult"
	PassengerTypeChild  = "child"
	PassengerTypeInfant = "infant"
)

type BookingDetail struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`
	BookingID uint `json:"booking_id" gorm:"not null"`
	PassengerTitle string    `json:"passenger_title" gorm:"size:10;not null"` // tuan, nyonya, nona
	PassengerName  string    `json:"passenger_name" gorm:"size:255;not null"`
	PassengerDOB   time.Time `json:"passenger_dob" gorm:"type:date;not null"`
	PassengerType  string    `json:"passenger_type" gorm:"size:20;not null"`
	Nationality    string    `json:"nationality" gorm:"size:50;not null"`
	PassportNumber *string    `json:"passport_number" gorm:"size:50"`
	IssuingCountry *string    `json:"issuing_country" gorm:"size:50"`
	ValidUntil     *time.Time `json:"valid_until" gorm:"type:date"`
	TicketNumber string `json:"ticket_number" gorm:"size:50;uniqueIndex;not null"`
	SeatClass string  `json:"seat_class" gorm:"size:50;not null"`
	Price     float64 `json:"price" gorm:"type:numeric(15,2);not null"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (BookingDetail) TableName() string {
	return "booking_details"
}