package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// SendOTPViaSMS sends an OTP to a mobile number using Fast2SMS
func SendOTPViaSMS(mobile string, otp string) error {
	apiKey := os.Getenv("FAST2SMS_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("Fast2SMS API key not found in environment variables")
	}

	// Construct Fast2SMS API URL with Query Parameters
	apiURL := fmt.Sprintf("https://www.fast2sms.com/dev/bulkV2?authorization=%s&route=otp&variables_values=%s&flash=0&numbers=%s",
		url.QueryEscape(apiKey),
		url.QueryEscape(otp),
		url.QueryEscape(mobile),
	)

	// Make API request
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Fast2SMS Response:", string(body))

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Fast2SMS API request failed with status code: %d", resp.StatusCode)
	}

	return nil
}
