package domain

import (
	"strings"
	"time"
)

// Work represents the abstract intellectual creation (e.g., "The Hobbit" by J.R.R. Tolkien).
type Work struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	Subtitle          string    `json:"subtitle,omitempty"`
	OriginalYear      *int      `json:"original_year,omitempty"`
	Description       string    `json:"description,omitempty"`
	SubjectTags       []string  `json:"subject_tags,omitempty"`
	OpenLibraryWorkID string    `json:"open_library_work_id,omitempty"`
	GoogleBooksID     string    `json:"google_books_id,omitempty"`
	CoverURL          string    `json:"cover_url,omitempty"`
	Authors           []Author  `json:"authors,omitempty"`
	Editions          []Edition `json:"editions,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Validate ensures the work adheres to business invariants.
func (w *Work) Validate() error {
	if strings.TrimSpace(w.Title) == "" {
		return ErrEmptyTitle
	}
	return nil
}

// IsUUID checks whether a given string is a valid standard RFC 4122 UUID (36 chars: 8-4-4-4-12 hex).
func IsUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}
