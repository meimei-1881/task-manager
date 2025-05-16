package models

import "gorm.io/gorm"

// Task struct ใช้เก็บข้อมูลงาน
type Task struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`

	CreatedByID  uint `json:"created_by_id"`  // ID ของคนสร้าง
	UpdatedByID  uint `json:"updated_by_id"`  // ID ของคนแก้ไขล่าสุด
	AssignedToID uint `json:"assigned_to_id"` // ID ของคนแก้ไขล่าสุด
}
