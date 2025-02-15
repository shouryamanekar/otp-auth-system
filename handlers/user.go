package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetCurrentUser retrieves user details from the JWT token
// @Summary Get current user details
// @Description Returns the authenticated user's mobile number
// @Tags User
// @Security BearerToken
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /user [get]
func GetCurrentUser(c *gin.Context) {
	mobile, exists := c.Get("mobile")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User details retrieved successfully",
		"mobile":  mobile,
	})
}
