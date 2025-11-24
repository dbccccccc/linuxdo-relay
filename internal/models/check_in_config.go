package models

import "time"

// CheckInConfig defines base reward / decay threshold per level.
type CheckInConfig struct {
	ID                   uint      `gorm:"primaryKey"`
	Level                int       `gorm:"uniqueIndex;not null"`
	BaseReward           int       `gorm:"not null"`
	DecayThreshold       int       `gorm:"not null"`
	MinMultiplierPercent int       `gorm:"not null;default:10"`
	CreatedAt            time.Time `gorm:"not null"`
	UpdatedAt            time.Time `gorm:"not null"`
}
