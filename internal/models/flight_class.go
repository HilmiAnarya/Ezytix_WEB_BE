package models

import "time"

// FlightClass Model (Kelas Kabin)
type FlightClass struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FlightID    uint      `json:"flight_id"`
	SeatClass   string    `json:"seat_class" gorm:"type:enum('economy', 'business', 'first_class');not null"` // Jenis kelas kabin
	Price       float64   `json:"price" gorm:"not null"` // Harga per kelas kabin
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

func (FlightClass) TableName() string {
	return "flight_classes"
}
