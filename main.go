package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"log"
	"os"
	"task-manager/database"
	"task-manager/handlers"
	"task-manager/internal/auth"
	"task-manager/internal/notification"
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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run database migrations
	if err := migrateDatabase(database.DB); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Initialize notification hub
	hub := notification.NewHub()
	go hub.Run() // Start hub in background

	// Initialize handlers
	wsHandler := notification.NewWSHandler(hub)
	notiHandler := handlers.NewNotificationHandler(database.DB, hub)
	taskHandler := handlers.NewTaskHandler(hub, notiHandler)

	// Setup router
	r := setupRouter(hub, wsHandler, taskHandler, notiHandler)

	// Start server
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func migrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Task{},
		&models.User{},
		&models.Notification{},
		&models.NotificationRecipient{},
	)
}

func setupRouter(
	hub *notification.Hub,
	wsHandler *notification.WSHandler,
	taskHandler *handlers.TaskHandler,
	notiHandler *handlers.NotificationHandler,

) *gin.Engine {
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// WebSocket route
	r.GET("/ws", func(c *gin.Context) {
		wsHandler.HandleConnection(c.Writer, c.Request)
	})

	// Public routes
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	// Authenticated routes
	api := r.Group("/api")
	api.Use(auth.GinAuthMiddleware())
	{
		// Task routes
		api.GET("/tasks", handlers.GetTasks)
		api.POST("/tasks", handlers.CreateTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.GET("/tasks/:id", handlers.GetTaskByID)
		api.DELETE("/tasks/:id", handlers.DeleteTask)

		// User routes
		api.GET("/users", handlers.GetUsers)

		// Notification routes
		api.GET("/notifications", handlers.GetNotifications)
		api.GET("/notifications/:id", handlers.GetNotificationByID)
		api.DELETE("/notifications/:id", handlers.DeleteNotification)
		api.PUT("/notifications/:id/mark-read", handlers.MarkNotificationAsRead)
		api.PUT("/notifications/mark-all-read", handlers.MarkAllAsRead)
	}

	return r
}
