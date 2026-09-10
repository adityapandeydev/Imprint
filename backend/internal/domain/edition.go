package domain

import (
	"strings"
	"time"
)

// BookFormat categorizes the physical or digital manifestation of an edition.
type BookFormat string

const (
	FormatHardcover  BookFormat = "HARDCOVER"
	FormatPaperback  BookFormat = "PAPERBACK"
	FormatMassMarket BookFormat = "MASS_MARKET"
	FormatEbook      BookFormat = "EBOOK"
	FormatAudiobook  BookFormat = "AUDIOBOOK"
	FormatUnknown    BookFormat = "UNKNOWN"
)

// ParseBookFormat safely converts a string representation into a valid BookFormat.
func ParseBookFormat(raw string) BookFormat {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	switch {
	case strings.Contains(normalized, "HARDCOVER") || strings.Contains(normalized, "HARDBACK"):
		return FormatHardcover
	case strings.Contains(normalized, "MASS"):
		return FormatMassMarket
	case strings.Contains(normalized, "PAPERBACK") || strings.Contains(normalized, "SOFTCOVER"):
		return FormatPaperback
	case strings.Contains(normalized, "EBOOK") || strings.Contains(normalized, "KINDLE") || strings.Contains(normalized, "EPUB"):
		return FormatEbook
	case strings.Contains(normalized, "AUDIO"):
		return FormatAudiobook
	default:
		return FormatUnknown
	}
}

// Edition represents a concrete published edition of a Work.
type Edition struct {
	ID                   string     `json:"id"`
	WorkID               string     `json:"work_id"`
	Title                string     `json:"title"`
	Publisher            string     `json:"publisher,omitempty"`
	PublicationDate      string     `json:"publication_date,omitempty"`
	PublicationYear      *int       `json:"publication_year,omitempty"`
	PageCount            *int       `json:"page_count,omitempty"`
	Language             string     `json:"language,omitempty"`
	Format               BookFormat `json:"format"`
	ISBN10               *string    `json:"isbn10,omitempty"`
	ISBN13               *string    `json:"isbn13,omitempty"`
	ASIN                 *string    `json:"asin,omitempty"`
	CoverURL             string     `json:"cover_url,omitempty"`
	Description          string     `json:"description,omitempty"`
	OpenLibraryEditionID string     `json:"open_library_edition_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
