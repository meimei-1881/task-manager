// auth/middleware.go
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
)

var JWTSecret = []byte(os.Getenv("JWT_SECRET")) // ควรเก็บใน environment variable จริงๆ

func GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := tokenParts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"details": err.Error(),
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims type"})
			return
		}

		// Debug: Log all claims
		//log.Printf("Token Claims: %+v", claims)

		// ดึง userID (รองรับทั้งรูปแบบ snake_case และ camelCase)
		var userID uint
		if val, exists := claims["user_id"]; exists {
			userID = uint(val.(float64))
		} else if val, exists := claims["userID"]; exists {
			userID = uint(val.(float64))
		} else {
			c.JSON(401, gin.H{"error": "User ID not found in token"})
			c.Abort()
			return
		}

		// ดึง username
		username, ok := claims["username"].(string)
		if !ok {
			c.JSON(401, gin.H{"error": "Username not found in token"})
			c.Abort()
			return
		}

		if ok {
			// ตั้งค่าใน context
			c.Set("userID", userID)
			c.Set("username", username)
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid user ID in claims"})
	}
}
