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
