package models

import "time"

type Airline struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	IATA      string    `gorm:"type:varchar(10);unique;not null" json:"IATA"` // IATA Code (GA, JT)
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	LogoURL   string    `gorm:"type:text" json:"logo_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}