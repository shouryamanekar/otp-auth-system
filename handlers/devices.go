package handlers

import (
	"net/http"
	"otp-auth-system/db"
	"otp-auth-system/utils"

	"github.com/gin-gonic/gin"
)

// DeviceRequest DeleteDeviceRequest defines the request body for removing a device
type DeviceRequest struct {
	DeviceFingerprint string `json:"device_fingerprint"`
}

// GetRegisteredDevices retrieves all registered devices for a user
// @Summary Get registered devices
// @Description Returns a list of devices where the user has logged in
// @Tags Devices
// @Security BearerToken
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/devices [get]
func GetRegisteredDevices(c *gin.Context) {
	mobile, exists := c.Get("mobile")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var devices []string
	err := db.DB.Select(&devices, "SELECT device_fingerprint FROM user_devices WHERE mobile = $1", mobile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registered devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

// RemoveRegisteredDevice deletes a specific registered device
// @Summary Remove a specific device
// @Description Deletes a registered device from the user's account
// @Tags Devices
// @Security BearerToken
// @Accept json
// @Produce json
// @Param request body handlers.DeviceRequest true "Device Fingerprint "
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /delete [delete]
func RemoveRegisteredDevice(c *gin.Context) {
	mobile, exists := c.Get("mobile")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var request struct {
		DeviceFingerprint string `json:"device_fingerprint"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Delete the specific device from user_devices table
	result, err := db.DB.Exec("DELETE FROM user_devices WHERE mobile = $1 AND device_fingerprint = $2", mobile, request.DeviceFingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove device"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device removed successfully"})
}

// RemoveAllOtherDevices removes all devices except the current one
// @Summary Remove all devices except current
// @Description Logs out all devices except the currently active one
// @Tags Devices
// @Security BearerToken
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/all [delete]
func RemoveAllOtherDevices(c *gin.Context) {
	mobile, exists := c.Get("mobile")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Generate the fingerprint for the current device
	currentFingerprint := utils.GenerateFingerprint(c.Request)

	// Ensure the current device is NOT removed
	result, err := db.DB.Exec("DELETE FROM user_devices WHERE mobile = $1 AND device_fingerprint != $2", mobile, currentFingerprint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove devices"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No other devices found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All other devices removed successfully, current device remains"})
}
