package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"otp-auth-system/cache"
	"otp-auth-system/db"
	"otp-auth-system/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// VerifyOTPRequest defines the request body for verifying OTP
type VerifyOTPRequest struct {
	Mobile string `json:"mobile"`
	OTP    string `json:"otp"`
}

// VerifyOTP verifies OTP for authentication
// @Summary Verify OTP
// @Description Confirms OTP and authenticates user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body handlers.VerifyOTPRequest true "User's mobile number and OTP"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /verify [post]
func VerifyOTP(c *gin.Context) {
	var request struct {
		Mobile string `json:"mobile"`
		OTP    string `json:"otp"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Retrieve OTP from Redis
	storedOTP, err := cache.RDB.Get(context.Background(), request.Mobile).Result()
	if err != nil || storedOTP != request.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	// OTP is correct, remove it from Redis
	cache.RDB.Del(context.Background(), request.Mobile)

	// Generate JWT token
	token, err := utils.GenerateJWT(request.Mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Generate current fingerprint
	currentFingerprint := utils.GenerateFingerprint(c.Request)

	// Check if fingerprint already exists
	var existingFingerprint string
	err = db.DB.Get(&existingFingerprint, "SELECT device_fingerprint FROM user_devices WHERE mobile = $1 AND device_fingerprint = $2", request.Mobile, currentFingerprint)

	// Determine which fingerprint to use for JWT storage
	var fingerprintToUse string

	if err == nil {
		// ✅ Existing fingerprint found → Use it
		fingerprintToUse = existingFingerprint
	} else if errors.Is(err, sql.ErrNoRows) {
		// ❌ No fingerprint found → Store the new fingerprint
		fingerprintToUse = currentFingerprint
		_, err := db.DB.Exec("INSERT INTO user_devices (mobile, device_fingerprint) VALUES ($1, $2)", request.Mobile, currentFingerprint)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store new device fingerprint"})
			return
		}
	} else {
		// ⚠️ Unexpected database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error while checking device fingerprint"})
		return
	}

	// 🔐 Store JWT token in Redis mapped to the chosen fingerprint
	tokenKey := fmt.Sprintf("device_token:%s:%s", request.Mobile, fingerprintToUse)
	err = cache.RDB.Set(context.Background(), tokenKey, token, 24*time.Hour).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified, login successful", "token": token})
}
