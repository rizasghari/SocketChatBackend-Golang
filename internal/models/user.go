package models

import (
	"gorm.io/gorm"
	"time"
)

// User represents a user in the application
type User struct {
	gorm.Model
	FirstName    string     `gorm:"not null" json:"first_name"`
	LastName     string     `gorm:"not null" json:"last_name"`
	ProfilePhoto *string    `json:"profile_photo"`
	Email        string     `gorm:"unique;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Password     string     `gorm:"-" json:"password"`
	IsOnline     bool       `gorm:"default:false" json:"is_online"`
	LastSeen     *time.Time `json:"last_seen"`
}