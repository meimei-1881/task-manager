package models

import "gorm.io/gorm"

// Task struct ใช้เก็บข้อมูลงาน
type Task struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status" gorm:"default:'todo'"`
	Priority    string `json:"priority" gorm:"default:'medium'"`

	CreatedBy   User `json:"created_by" gorm:"foreignKey:CreatedByID"`
	CreatedByID uint `json:"created_by_id"`

	UpdatedBy   User `json:"updated_by" gorm:"foreignKey:UpdatedByID"`
	UpdatedByID uint `json:"updated_by_id"`

	AssignedTo   User `json:"assigned_to" gorm:"foreignKey:AssignedToID"`
	AssignedToID uint `json:"assigned_to_id"`
}
