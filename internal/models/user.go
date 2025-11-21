package models

import "time"

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	FullName  string    `json:"full_name" gorm:"size:255;not null"`

	// NEW FIELD: USERNAME
	Username  string    `json:"username" gorm:"size:16;uniqueIndex;not null"`

	Email     string    `json:"email" gorm:"size:255;uniqueIndex;not null"`
	Phone     string    `json:"phone" gorm:"size:20;uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`

	Role      UserRole  `json:"role" gorm:"type:enum('customer','admin');default:'customer'"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
