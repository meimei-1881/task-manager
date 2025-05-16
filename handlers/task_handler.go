package handlers

import (
	"github.com/gin-gonic/gin"
	"task-manager/database"
	"task-manager/models"
)

var err error

func GetTasks(c *gin.Context) {
	var tasks []models.Task
	if err := database.DB.Find(&tasks).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	c.JSON(200, tasks)
}

// ฟังก์ชันสำหรับสร้าง task ใหม่ในฐานข้อมูล
func CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// ดึง userID จาก JWT (หลังจากผ่าน AuthMiddleware)
	userID := c.MustGet("userID").(uint)
	task.CreatedByID = userID
	task.UpdatedByID = userID

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(200, gin.H{"task": task})
}

func UpdateTask(c *gin.Context) {
	var input struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Status       string `json:"status"`
		AssignedToID uint   `json:"assigned_to_id"`
	}

	// Bind JSON จาก Request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// ดึง userID ของคนที่กำลังแก้ไข (จาก JWT)
	updatedByID := c.MustGet("userID").(uint)

	// หา Task ที่ต้องการอัปเดต
	var task models.Task
	if err := database.DB.First(&task, c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "Task not found"})
		return
	}

	// อัปเดตค่า
	database.DB.Model(&task).Updates(models.Task{
		Name:         input.Name,
		Description:  input.Description,
		Status:       input.Status,
		AssignedToID: input.AssignedToID, // <-- อัปเดต AssignedToID
		UpdatedByID:  updatedByID,        // <-- บันทึกคนที่แก้ไขล่าสุด
	})

	c.JSON(200, gin.H{"task": task})
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
