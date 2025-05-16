package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"task-manager/auth"
	"task-manager/database"
	"task-manager/handlers"
	"task-manager/models"
)

//
//var db *gorm.DB
//var err error

//// เชื่อมต่อกับฐานข้อมูล
//func init() {
//	db, err = gorm.Open(sqlite.Open("tasks.db"), &gorm.Config{})
//	if err != nil {
//		panic("failed to connect to database")
//	}
//
//	// เรียกใช้ Migrate เพื่อสร้างตาราง Task
//	if err := models.Migrate(db); err != nil {
//		panic("failed to migrate database")
//	}
//}

func main() {
	// เชื่อมต่อ Database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate โมเดล User (สร้างตารางอัตโนมัติ)
	if err := database.DB.AutoMigrate(&models.Task{}, &models.User{}); err != nil { // <-- ต้องมีบรรทัดนี้
		log.Fatal("Failed to migrate database:", err)
	}

	// โหลดค่าจาก .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("API_PORT")

	r := gin.Default()
	// ตั้งค่า CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},                   // กำหนด URL ที่อนุญาตให้เชื่อมต่อ
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},            // กำหนดวิธีการ HTTP ที่อนุญาต
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"}, // ⚠️ ต้องมี Authorization
		AllowCredentials: true,                                                // ให้อนุญาต Cookies ในคำร้อง
	}))

	r.POST("/register", handlers.Register) // <-- เพิ่มบรรทัดนี้
	r.POST("/login", handlers.Login)

	api := r.Group("/api")

	api.Use(auth.GinAuthMiddleware()) // <-- เรียกใช้แบบนี้
	{
		api.GET("/tasks", handlers.GetTasks)
		api.POST("/tasks", handlers.CreateTask)
		api.PUT("/tasks/:id", handlers.UpdateTask)
		api.DELETE("/tasks/:id", handlers.DeleteTask)
	}

	if port == "" {
		port = "8080" // ค่า default เมื่อไม่ได้ตั้งค่า API_PORT ใน environment
	}
	r.Run(":" + port) // ต้องมี colon (:) นำหน้าหมายเลขพอร์ต}
}

//func getTasks(c *gin.Context) {
//	c.JSON(200, gin.H{
//		"tasks": []string{"task1", "task2"},
//	})
//}
