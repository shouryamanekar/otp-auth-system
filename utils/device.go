package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// GenerateFingerprint creates a unique hash for a device
func GenerateFingerprint(r *http.Request) string {
	userAgent := r.Header.Get("User-Agent")
	ip := r.RemoteAddr

	rawData := userAgent + ip
	hash := sha256.Sum256([]byte(rawData))

	return hex.EncodeToString(hash[:])
}
