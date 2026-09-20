package domain

import (
	"time"
)

// ProfileVisibility controls how a reader's library is accessed publicly.
type ProfileVisibility string

const (
	VisibilityPublic   ProfileVisibility = "PUBLIC"
	VisibilityUnlisted ProfileVisibility = "UNLISTED"
	VisibilityPrivate  ProfileVisibility = "PRIVATE"
)

// User represents an account within Imprint.
type User struct {
	ID                string            `json:"id"`
	Email             string            `json:"email"`
	Username          string            `json:"username"`
	DisplayName       string            `json:"display_name"`
	PasswordHash      string            `json:"-"` // Never serialized to JSON
	ProfileVisibility ProfileVisibility `json:"profile_visibility"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}
