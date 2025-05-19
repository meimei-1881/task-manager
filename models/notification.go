package models

import (
	"time"
)

type Notification struct {
	ID         uint                    `json:"id" gorm:"primaryKey"`
	Message    string                  `json:"message"`
	Type       string                  `json:"type"`
	TaskID     uint                    `json:"task_id"`
	CreatedAt  time.Time               `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	Recipients []NotificationRecipient `gorm:"foreignKey:NotificationID"`
}

type NotificationRecipient struct {
	NotificationID uint `json:"notification_id" gorm:"primaryKey"`
	UserID         uint `json:"user_id" gorm:"primaryKey"`
	Read           bool `json:"read" gorm:"default:false"`
}

type NotificationWithReadStatus struct {
	ID        uint      `json:"id"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	TaskID    uint      `json:"task_id"`
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"` // เพิ่มฟิลด์นี้แทน Recipients ทั้งหมด
}
