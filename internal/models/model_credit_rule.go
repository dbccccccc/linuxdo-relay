package models

import "time"

// ModelCreditRule defines per-model credit cost configuration.
type ModelCreditRule struct {
	ID           uint      `gorm:"primaryKey"`
	ModelPattern string    `gorm:"size:128;not null"`
	CreditCost   int       `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}
