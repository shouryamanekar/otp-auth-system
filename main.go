package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	_ "net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"otp-auth-system/cache"
	"otp-auth-system/db"
	"otp-auth-system/handlers"
	"otp-auth-system/middleware"

	_ "otp-auth-system/docs" // Import Swagger Docs
)

// @title OTP Authentication API
// @version 1.0
// @securityDefinitions.apikey BearerToken
// @in header
// @name Authorization
// @description Use JWT token to authorize API requests.
// @host localhost:8080
// @BasePath /
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}

	// Determine environment mode
	env := os.Getenv("ENV")
	if env == "" {
		env = "development" // Default to development mode
	}

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db.InitDB()
	defer func(DB *sqlx.DB) {
		err := DB.Close()
		if err != nil {
			log.Fatalf("Error closing database: %v", err)
		}
	}(db.DB)

	// Initialize Redis
	cache.InitRedis()
	defer func(RDB *redis.Client) {
		err := RDB.Close()
		if err != nil {
			log.Fatalf("Error closing Redis: %v", err)
		}
	}(cache.RDB)

	// Periodically clean up expired tokens
	go func() {
		for {
			time.Sleep(1 * time.Hour) // Run every hour
			cache.RemoveExpiredTokens()
		}
	}()

	// Set up router
	router := gin.Default()

	// Enable CORS for all origins
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Restrict Trusted Proxies
	err := router.SetTrustedProxies(nil)
	if err != nil {
		log.Fatalf("Error Restricted Proxies: %v", err)
	}

	// Enable Swagger UI only in non-production environments
	if env != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Routes
	router.POST("/register", handlers.RegisterUser) // Register new user
	router.POST("/login", handlers.LoginUser)       // Generate OTP for login
	router.POST("/verify", handlers.VerifyOTP)      // Verify OTP and authenticate user
	router.POST("/resend-otp", handlers.ResendOTP)

	// Protected Route (Requires JWT)
	protected := router.Group("/").Use(middleware.AuthMiddleware())

	protected.GET("/user", handlers.GetCurrentUser)                  // Get logged-in user details
	protected.GET("/user/devices", handlers.GetRegisteredDevices)    // Get logged-in user details
	protected.DELETE("/device", handlers.RemoveRegisteredDevice)     // Remove a specific device
	protected.DELETE("/devices/all", handlers.RemoveAllOtherDevices) // Remove all devices except current one
	protected.POST("/logout", handlers.Logout)                       // Logout from current device
	protected.POST("/logout/all", handlers.LogoutAll)                // Logout from all devices

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	fmt.Printf("Server running on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
