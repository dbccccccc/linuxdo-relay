package models

import "time"

// LoginLog records a LinuxDo OAuth login for audit and security purposes.
type LoginLog struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	IPAddress string    `gorm:"size:64"`
	UserAgent string    `gorm:"size:255"`
	CreatedAt time.Time `gorm:"not null;index"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}
