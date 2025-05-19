package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"task-manager/database"
	"task-manager/internal/notification"
	"task-manager/models"
	"time"
)

var err error

type TaskHandler struct {
	notificationHub     *notification.Hub // เชื่อมโยงกับ NotificationHub
	notificationHandler *NotificationHandler
}

func NewTaskHandler(notificationHub *notification.Hub, notificationHandler *NotificationHandler) *TaskHandler {
	return &TaskHandler{
		notificationHub:     notificationHub,
		notificationHandler: notificationHandler,
	}
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var input models.Task

	// Bind JSON จาก Request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// ดึง userID ของคนที่กำลังแก้ไข (จาก JWT)
	updatedByID := c.MustGet("userID").(uint)
	updatedByUsername := c.MustGet("username")

	// หา Task ที่ต้องการอัปเดต
	var task models.Task
	if err := database.DB.Preload("AssignedTo").First(&task, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	// อัปเดตค่า
	if err := database.DB.Model(&task).Omit("CreatedAt").Updates(models.Task{
		Name:         input.Name,
		Description:  input.Description,
		Status:       input.Status,
		AssignedToID: input.AssignedToID,
		UpdatedByID:  updatedByID,
		Priority:     input.Priority,
	}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update task"})
		return
	}

	// สร้าง Notification หลัก
	notification := models.Notification{
		Message:   fmt.Sprintf("Task '%s' has been updated by %s", task.Name, c.MustGet("username").(string)),
		Type:      "task_update",
		TaskID:    task.ID,
		CreatedAt: time.Now(),
	}

	// บันทึก Notification หลักลงฐานข้อมูล
	if err := database.DB.Create(&notification).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create notification"})
		return
	}

	// ดึงรายชื่อ User ทั้งหมดที่ควรได้รับ Notification
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch users"})
		return
	}

	message := fmt.Sprintf("Task '%s' has been updated by %s", task.Name, updatedByUsername)
	h.notificationHandler.SendTaskUpdateNotification(&task, "task_update", message)

	c.JSON(http.StatusOK, gin.H{"task": task})
}

func GetTasks(c *gin.Context) {
	var tasks []models.Task
	if err := database.DB.Find(&tasks).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	c.JSON(200, tasks)
}

// CreateTask handles task creation
func CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate priority
	validPriorities := map[string]bool{"low": true, "medium": true, "high": true}
	if !validPriorities[task.Priority] {
		task.Priority = "medium" // Default value
	}

	// Set created by
	userID, _ := c.Get("userID")
	task.CreatedByID = userID.(uint)

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// DELETE /api/tasks/:id
func DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := database.DB.Where("id = ?", id).Delete(&models.Task{}).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(200, gin.H{"message": "Task deleted successfully"})
}

func GetTaskByID(c *gin.Context) {
	id := c.Param("id")

	var tasks models.Task
	if err := database.DB.Where("id = ?", id).Find(&tasks).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	c.JSON(200, tasks)

}
