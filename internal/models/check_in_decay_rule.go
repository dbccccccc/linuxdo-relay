package models

import "time"

// CheckInDecayRule represents a global multiplier threshold.
type CheckInDecayRule struct {
	ID                uint      `gorm:"primaryKey"`
	Threshold         int       `gorm:"not null"`
	MultiplierPercent int       `gorm:"not null"`
	SortOrder         int       `gorm:"default:0"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
}
