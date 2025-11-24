package models

import "time"

// QuotaRule defines per-level, per-model-pattern request limits.
type QuotaRule struct {
	ID            uint      `gorm:"primaryKey"`
	Level         int       `gorm:"not null"`
	ModelPattern  string    `gorm:"size:64;not null"`
	MaxRequests   int       `gorm:"not null"`
	WindowSeconds int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}
