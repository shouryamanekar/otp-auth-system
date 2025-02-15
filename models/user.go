package models

import "github.com/google/uuid"

// User struct represents a user in the system
type User struct {
	ID                uuid.UUID `db:"id" json:"id"`
	Mobile            string    `db:"mobile" json:"mobile"`
	DeviceFingerprint string    `db:"device_fingerprint" json:"device_fingerprint"`
}
