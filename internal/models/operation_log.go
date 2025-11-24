package models

import "time"

// OperationLog records a user-initiated action in the console, such as
// regenerating an API key.
type OperationLog struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"not null;index"`
	OperationType string    `gorm:"size:64;not null"`
	Details       string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null;index"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
