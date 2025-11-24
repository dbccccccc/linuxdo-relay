package models

import "time"

const (
	ChannelStatusEn  = "enabled"
	ChannelStatusDis = "disabled"
)

// Channel represents an upstream channel configuration for new-api.
// All channels are logically the same type (new-api); fields Type/Priority
// have been removed and each channel is distinguished only by base_url,
// api_key, supported models, and status.
type Channel struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:64;not null"`
	BaseURL   string    `gorm:"column:base_url;not null"`
	APIKey    string    `gorm:"column:api_key;not null"`
	Models    string    `gorm:"type:jsonb;not null"`
	Status    string    `gorm:"size:16;not null;default:'enabled'"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
