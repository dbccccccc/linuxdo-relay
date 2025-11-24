package models

import "time"

// UserRole defines user role.
const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

// UserStatus defines status.
const (
	UserStatusNormal   = "normal"
	UserStatusDisabled = "disabled"
)

// User represents a local user mapped from LinuxDo account.
type User struct {
	ID              uint       `gorm:"primaryKey"`
	LinuxDoUserID   int64      `gorm:"column:linuxdo_user_id;uniqueIndex;not null"`
	LinuxDoUsername string     `gorm:"column:linuxdo_username;size:255;not null"`
	Role            string     `gorm:"size:16;not null"`
	Level           int        `gorm:"not null;default:1"`
	Status          string     `gorm:"size:16;not null;default:'normal'"`
	Credits         int        `gorm:"not null;default:0"`
	APIKeyHash      string     `gorm:"column:api_key_hash;size:128"`
	APIKeyCreatedAt *time.Time `gorm:"column:api_key_created_at"`
	CreatedAt       time.Time  `gorm:"not null"`
	UpdatedAt       time.Time  `gorm:"not null"`
}
