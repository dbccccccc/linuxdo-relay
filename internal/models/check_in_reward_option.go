package models

import "time"

// CheckInRewardOption defines a single wheel slice configuration.
type CheckInRewardOption struct {
	ID          uint      `gorm:"primaryKey"`
	Label       string    `gorm:"size:64;not null"`
	Credits     int       `gorm:"not null"`
	Probability int       `gorm:"not null"`
	Color       string    `gorm:"size:24"`
	SortOrder   int       `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
