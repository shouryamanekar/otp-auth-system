package handlers

import (
	"context"
	"fmt"
	"net/http"
	"otp-auth-system/cache"
	"otp-auth-system/db"

	"github.com/gin-gonic/gin"
)

// Logout logs out the user from the current device
// @Summary Logout from current device
// @Description Invalidates JWT token from Redis for the current device
// @Tags Authentication
// @Security BearerToken
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /logout [post]
func Logout(c *gin.Context) {
	// Get token from Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		return
	}

	// Blacklist the token
	err := cache.BlacklistToken(tokenString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Remove token from Redis session
	cache.RDB.Del(context.Background(), "device_token:"+tokenString)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// LogoutAll logs out the user from all devices
// @Summary Logout from all devices
// @Description Invalidates JWT tokens for all devices of the user
// @Tags Authentication
// @Security BearerToken
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /logout/all [post]
func LogoutAll(c *gin.Context) {
	mobile, exists := c.Get("mobile")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Retrieve user's device fingerprints from the database
	var deviceFingerprints []string
	err := db.DB.Select(&deviceFingerprints, "SELECT device_fingerprint FROM user_devices WHERE mobile = $1", mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user devices"})
		return
	}

	// Blacklist only the JWTs associated with these devices
	for _, device := range deviceFingerprints {
		tokenKey := fmt.Sprintf("device_token:%s:%s", mobile, device)
		token, err := cache.RDB.Get(context.Background(), tokenKey).Result()
		if err == nil && token != "" {
			cache.BlacklistToken(token)
			cache.RDB.Del(context.Background(), tokenKey) // Remove device-token mapping
		}
	}

	// Remove all device records for the user
	_, err = db.DB.Exec("DELETE FROM user_devices WHERE mobile = $1", mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices successfully"})
}
