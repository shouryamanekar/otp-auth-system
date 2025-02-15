package middleware

import (
	"fmt"
	"net/http"
	"otp-auth-system/cache"
	"otp-auth-system/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for a valid JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Check if token is blacklisted
		isBlacklisted, _ := cache.IsTokenBlacklisted(tokenString)
		if isBlacklisted {
			fmt.Println("🚨 Blacklisted token detected:", tokenString) // Debugging log
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is invalid or expired"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store user information in the request context
		c.Set("mobile", claims.Mobile)

		c.Next()
	}
}
