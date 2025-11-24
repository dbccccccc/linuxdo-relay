package models

import "time"

// CreditTransaction keeps audit logs for every credit balance change.
type CreditTransaction struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	Delta     int       `gorm:"not null"`
	Reason    string    `gorm:"size:64;not null"`
	Status    string    `gorm:"size:16;not null"`
	ModelName string    `gorm:"size:128"`
	RequestID string    `gorm:"size:64"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
