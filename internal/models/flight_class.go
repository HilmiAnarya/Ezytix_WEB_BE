package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type FlightClass struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FlightID    uint      `json:"flight_id"`
	SeatClass   string    `json:"seat_class" gorm:"type:enum('economy', 'business', 'first_class');not null"`
	ClassCode  string          `json:"class_code"`          // [BARU] Sub-kelas (Contoh: I9, Y, H)
	Price      decimal.Decimal `json:"price" gorm:"type:numeric(15,2);not null"`
	TotalSeats  int       `json:"total_seats" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func (FlightClass) TableName() string {
	return "flight_classes"
}
