package models

import "time"

// APILog records a single API call made through the relay.
// Status is a simple "success" / "fail" flag, while StatusCode stores the
// upstream HTTP status code returned by new-api.
type APILog struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null;index"`
	Model        string    `gorm:"size:128;not null"`
	Status       string    `gorm:"size:32;not null"`
	StatusCode   int       `gorm:"not null"`
	ErrorMessage string    `gorm:"type:text"`
	IPAddress    string    `gorm:"size:64"`
	CreatedAt    time.Time `gorm:"not null;index"`
}

func (APILog) TableName() string {
	return "api_logs"
}
