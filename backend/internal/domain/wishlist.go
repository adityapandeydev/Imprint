package domain

import (
	"strings"
	"time"
)

// ReadingStatus defines the user's progress through a book.
type ReadingStatus string

const (
	StatusWantToRead         ReadingStatus = "WANT_TO_READ"
	StatusCurrentlyReading   ReadingStatus = "CURRENTLY_READING"
	StatusFinished           ReadingStatus = "FINISHED"
	StatusAbandoned          ReadingStatus = "ABANDONED"
)

// ValidateReadingStatus verifies that a status string matches an allowed domain status.
func ValidateReadingStatus(raw string) (ReadingStatus, error) {
	switch ReadingStatus(strings.ToUpper(strings.TrimSpace(raw))) {
	case StatusWantToRead:
		return StatusWantToRead, nil
	case StatusCurrentlyReading:
		return StatusCurrentlyReading, nil
	case StatusFinished:
		return StatusFinished, nil
	case StatusAbandoned:
		return StatusAbandoned, nil
	default:
		return "", ErrInvalidStatus
	}
}

// WishlistItem represents an entry in a user's reading collection / wishlist.
type WishlistItem struct {
	ID         string        `json:"id"`
	UserID     string        `json:"user_id"`
	WorkID     string        `json:"work_id"`
	Work       *Work         `json:"work,omitempty"`
	EditionID  *string       `json:"edition_id,omitempty"`
	Edition    *Edition      `json:"edition,omitempty"`
	Status     ReadingStatus `json:"status"`
	Priority   int           `json:"priority"` // 1 (lowest) to 5 (highest), default 3
	Rating     *int          `json:"rating,omitempty"` // 1 to 5 stars
	Notes      string        `json:"notes,omitempty"`
	StartedAt  *time.Time    `json:"started_at,omitempty"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// Validate checks business rules for a wishlist item.
func (item *WishlistItem) Validate() error {
	if _, err := ValidateReadingStatus(string(item.Status)); err != nil {
		return err
	}
	if item.Priority < 1 || item.Priority > 5 {
		return ErrInvalidPriority
	}
	if item.Rating != nil {
		if *item.Rating < 1 || *item.Rating > 5 {
			return ErrInvalidRating
		}
	}
	return nil
}
