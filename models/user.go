package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName      string    `gorm:"not null" json:"first_name"`
	LastName       string    `gorm:"not null" json:"last_name"`
	Username       string    `gorm:"unique;not null" json:"username"`
	Email          string    `gorm:"unique;not null" json:"email"`
	PhoneNumber    string    `gorm:"unique;not null" json:"phone_number"`
	Password       string    `gorm:"not null" json:"password"`
	Token          string    `json:"token"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
}

