package handlers

import (
	"context"
	"net/http"
	"otp-auth-system/cache"
	"otp-auth-system/db"
	"otp-auth-system/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// OTP request limit settings
const otpRequestLimit = 6          // Maximum OTP requests per hour (including login, register, resend)
const otpBlockDuration = time.Hour // Duration before reset

// Function to check OTP rate limit in Redis
func isRateLimited(mobile string) bool {
	requestCountKey := "otp_requests:" + mobile
	requests, _ := cache.RDB.Get(context.Background(), requestCountKey).Int()

	return requests >= otpRequestLimit
}

// Function to increase OTP request count in Redis
func incrementOTPRequestCount(mobile string) {
	requestCountKey := "otp_requests:" + mobile
	cache.RDB.Incr(context.Background(), requestCountKey)
	cache.RDB.Expire(context.Background(), requestCountKey, otpBlockDuration)
}

// RegisterRequest defines the request body for user registration
type RegisterRequest struct {
	Mobile string `json:"mobile"`
}

// LoginRequest defines the request body for user login
type LoginRequest struct {
	Mobile string `json:"mobile"`
}

// ResendOTPRequest defines the request body for resending OTP
type ResendOTPRequest struct {
	Mobile string `json:"mobile"`
}

// RegisterUser registers a new user
// @Summary Register a new user
// @Description Registers a new user and sends OTP via SMS
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body handlers.RegisterRequest true "User's mobile number"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
func RegisterUser(c *gin.Context) {
	var request struct {
		Mobile string `json:"mobile"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if user already exists
	var exists bool
	err := db.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE mobile = $1)", request.Mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User already registered. Please log in."})
		return
	}

	// Store user in DB (if new)
	_, err = db.DB.Exec("INSERT INTO users (mobile) VALUES ($1) ON CONFLICT (mobile) DO NOTHING", request.Mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

// LoginUser logs in a user via OTP
// @Summary Login user with OTP
// @Description Sends OTP to the registered mobile number for authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body handlers.LoginRequest true "User's mobile number"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
func LoginUser(c *gin.Context) {
	var request struct {
		Mobile string `json:"mobile"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if user exists
	var exists bool
	err := db.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE mobile = $1)", request.Mobile)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check rate limit
	if isRateLimited(request.Mobile) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many OTP requests. Try again later."})
		return
	}

	// Generate OTP
	otp := utils.GenerateOTP()

	// Store OTP in Redis with a 5-minute expiration
	err = cache.RDB.Set(context.Background(), request.Mobile, otp, 5*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store OTP"})
		return
	}

	// Increment OTP request count
	incrementOTPRequestCount(request.Mobile)

	// Send OTP via SMS
	if err := utils.SendOTPViaSMS(request.Mobile, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP via SMS"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent via SMS"})
}

// ResendOTP sends a new OTP if the previous one expired
// @Summary Resend OTP
// @Description Requests a new OTP if the previous one expired
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body handlers.ResendOTPRequest true "User's mobile number"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /resend-otp [post]
func ResendOTP(c *gin.Context) {
	var request struct {
		Mobile string `json:"mobile"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if the user exists
	var exists bool
	err := db.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE mobile = $1)", request.Mobile)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check rate limit
	if isRateLimited(request.Mobile) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many OTP requests. Try again later."})
		return
	}

	// Generate a new OTP
	newOTP := utils.GenerateOTP()
	cache.RDB.Set(context.Background(), request.Mobile, newOTP, 5*time.Minute)

	// Increment OTP request count
	incrementOTPRequestCount(request.Mobile)

	// Send OTP via SMS
	if err := utils.SendOTPViaSMS(request.Mobile, newOTP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP via SMS"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "New OTP sent via SMS"})
}
