package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"log"
	"os"
	"time"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET")) // ควรเก็บใน environment variable จริงๆ

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint) (string, error) {
	log.Println("jwt ", jwtSecret)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "task-manager",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
