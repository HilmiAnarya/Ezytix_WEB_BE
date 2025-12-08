package models

import (
	"time"
)

type Airport struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement"`

	Code        string `json:"code" gorm:"size:3;uniqueIndex;not null"`
	CityName    string `json:"city_name" gorm:"size:100;not null"`
	AirportName string `json:"airport_name" gorm:"size:150;not null"`
	Country     string `json:"country" gorm:"size:100;not null"`

	CreatedAt 	time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt 	time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (Airport) TableName() string {
	return "airports"
}
