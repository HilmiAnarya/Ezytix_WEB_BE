package models

import "time"

type UserOTP struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	User      User      `json:"-" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	OTPCode   string    `json:"otp_code" gorm:"size:6;not null"`
	ExpiredAt time.Time `json:"expired_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserOTP) TableName() string {
	return "user_otps"
}