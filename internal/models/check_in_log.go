package models

import "time"

// CheckInLog records a user's daily check-in history.
type CheckInLog struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"not null;index"`
	CheckInDate   time.Time `gorm:"type:date;not null"`
	EarnedCredits int       `gorm:"not null"`
	Streak        int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null"`
}
