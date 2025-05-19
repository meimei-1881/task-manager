// auth/jwt.go
package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
)

var JWT_SECRET = []byte(os.Getenv("JWT_SECRET")) // อ่านจาก .env
//
//func GenerateToken(user models.User) (string, error) {
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//		"user_id": user.ID,
//		"exp":     time.Now().Add(time.Hour * 24).Unix(), // หมดอายุใน 24 ชั่วโมง
//	})
//	return token.SignedString(JWT_SECRET)
//}

func ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JWT_SECRET, nil
	})
}
