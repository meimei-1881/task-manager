package handlers

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"task-manager/auth"
	"task-manager/database"
	"task-manager/models"
)

// ค่า Secret Key (ควรเก็บใน .env)
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// ใน login handler
func Login(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// หา user ใน database
	var dbUser models.User
	if err := database.DB.Where("username = ?", user.Username).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username or password is incorrect"})
		return
	}

	// ตรวจสอบ password
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username or password is incorrect"})
		return
	}

	// สร้าง token
	token, err := auth.GenerateToken(dbUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user_id": dbUser.ID,
	})
}

func Register(c *gin.Context) {
	log.Printf("register func")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Debug: Log username ที่พยายามสมัคร
	log.Printf("Attempting to register username: %s", user.Username) // <-- เพิ่มบรรทัดนี้

	// ตรวจสอบว่า username ซ้ำหรือไม่
	var existingUser models.User
	result := database.DB.Where("username = ?", user.Username).First(&existingUser)

	// Debug: Log ผลลัพธ์การค้นหา
	log.Printf("DB search result: %+v, error: %v", existingUser, result.Error) // <-- เพิ่มบรรทัดนี้

	if result.Error == nil {
		// ถ้าไม่ error แสดงว่าพบ user นี้อยู่แล้ว
		log.Printf("Username already exists: %s", user.Username)
		c.JSON(409, gin.H{"error": "Username already exists"})
		return
	}

	// ถ้า username ไม่ซ้ำ, ทำการ hash password และสร้าง user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Password hashing error: %v", err)
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	if err := database.DB.Create(&user).Error; err != nil {
		log.Printf("Database create error: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}

	log.Printf("User registered successfully: %s", user.Username)
	c.JSON(200, gin.H{"message": "User registered successfully"})
}
