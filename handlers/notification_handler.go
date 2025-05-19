package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"sync"
	"task-manager/database"
	"task-manager/internal/notification"
	"task-manager/models"
	"time"
)

type NotificationHandler struct {
	db              *gorm.DB
	notificationHub *notification.Hub // ควรส่งมาจากส่วนที่เรียกใช้
}

func NewNotificationHandler(db *gorm.DB, hub *notification.Hub) *NotificationHandler {
	return &NotificationHandler{
		db:              db,
		notificationHub: hub,
	}
}

func (h *NotificationHandler) SendTaskUpdateNotification(task *models.Task, message, typeAction string) {
	go func() {
		// 1. สร้าง Notification หลัก
		notification := models.Notification{
			Message:   message,
			Type:      typeAction,
			TaskID:    task.ID,
			CreatedAt: time.Now(),
		}

		// ใช้ DB instance จาก Handler แทน global database.DB
		if err := h.db.Create(&notification).Error; err != nil {
			log.Printf("Failed to create notification: %v", err)
			return
		}

		// 2. ส่งให้ผู้ใช้ทั้งหมดแบบ Batch
		if err := h.createRecipients(notification.ID); err != nil {
			log.Printf("Failed to create recipients: %v", err)
			return
		}

		// 3. Broadcast
		if h.notificationHub != nil {
			h.notificationHub.Broadcast(models.Notification{
				ID:        notification.ID,
				Message:   notification.Message,
				Type:      notification.Type,
				TaskID:    notification.TaskID,
				CreatedAt: time.Time{},
			})
		}
	}()
}

func (h *NotificationHandler) createRecipients(notificationID uint) error {
	// ใช้ Batch Insert แทนการ loop สร้างทีละ record
	return h.db.Exec(`
		INSERT INTO notification_recipients (notification_id, user_id, read)
		SELECT ? AS notification_id, id AS user_id, false AS read
		FROM users
	`, notificationID).Error
}

// สร้าง Notification ใหม่ส่งให้ทุกคน
func CreateNotificationForAllUsers(db *gorm.DB, message string, taskID uint) error {
	// 1. สร้าง Notification หลัก
	notification := models.Notification{
		Message: message,
		TaskID:  taskID,
		Type:    "task_update",
	}
	if err := db.Create(&notification).Error; err != nil {
		return err
	}

	// 2. หารายชื่อ User ทั้งหมด
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	// 3. สร้าง Recipient สำหรับทุกคน
	var recipients []models.NotificationRecipient
	for _, user := range users {
		recipients = append(recipients, models.NotificationRecipient{
			NotificationID: notification.ID,
			UserID:         user.ID,
			Read:           false,
		})
	}

	return db.Create(&recipients).Error
}

// ดึง Notification ของ User หนึ่งคน
func GetNotifications(c *gin.Context) {
	log.Println("GetNotifications")

	userID := c.MustGet("userID").(uint)

	notifications, errNoti := GetUserNotifications(userID)
	if errNoti != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	// ส่งข้อมูล notifications เป็น Array
	c.JSON(200, notifications)
}

func GetNotificationByID(c *gin.Context) {
	id := c.Param("id")
	var notification models.Notification
	if err := database.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notification": notification})
}

func DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	var notification models.Notification
	if err := database.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// ลบ Notification
	if err := database.DB.Delete(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// MARK ALL AS READ สำหรับ user นั้นๆ
func MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	log.Println("MarkAllAsRead")

	var wg sync.WaitGroup
	// ใช้ Go Routine เพื่ออัปเดตข้อมูลในฐานข้อมูล
	wg.Add(1)
	go func() {
		defer wg.Done()
		// อัปเดตข้อมูล NotificationRecipient ว่าอ่านแล้ว
		if err := database.DB.Model(&models.NotificationRecipient{}).
			Where("user_id = ?", userID).
			Update("read", true).Error; err != nil {
			log.Println("Error marking notifications as read:", err)
		}
	}()

	// รอให้ Go Routine ทั้งหมดทำงานเสร็จ
	wg.Wait()

	// ส่ง Response กลับไปที่ client
	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// ดึง Notification ของ User พร้อมสถานะ
func GetUserNotifications(userID uint) ([]models.NotificationWithReadStatus, error) {
	var notifications []models.NotificationWithReadStatus

	err := database.DB.
		Table("notifications").
		Select("notifications.*, notification_recipients.read").
		Joins("JOIN notification_recipients ON notification_recipients.notification_id = notifications.id").
		Where("notification_recipients.user_id = ?", userID).
		Order("notifications.created_at DESC").
		Scan(&notifications).Error

	if err != nil {
		return nil, err
	}

	return notifications, nil
}

// ตั้งค่า Notification เป็นอ่านแล้ว
func MarkAsRead(notificationID, userID uint) error {
	return database.DB.Model(&models.NotificationRecipient{}).
		Where("notification_id = ? AND user_id = ?", notificationID, userID).
		Update("read", true).Error
}

// นับจำนวน Notification ที่ยังไม่ได้อ่าน
func CountUnreadNotifications(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&models.NotificationRecipient{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// ในไฟล์ handlers/notification.go
func MarkNotificationAsRead(c *gin.Context) {
	// ดึง userID จาก context (ถ้าใช้ JWT)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// แปลง notificationID จาก URL parameter
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid notification ID"})
		return
	}

	// เรียกใช้ฟังก์ชันเดิมของคุณ
	err = MarkAsRead(uint(notificationID), userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "marked as read"})
}
