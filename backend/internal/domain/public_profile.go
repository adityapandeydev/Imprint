package domain

import "time"

// PublicProfile represents a publicly shareable reader profile.
type PublicProfile struct {
	Username          string            `json:"username"`
	DisplayName       string            `json:"display_name"`
	ProfileVisibility ProfileVisibility `json:"profile_visibility"`
	MemberSince       time.Time         `json:"member_since"`
	Stats             *ReadingStats     `json:"stats,omitempty"`
	Shelves           []TagCount        `json:"shelves"`
}
