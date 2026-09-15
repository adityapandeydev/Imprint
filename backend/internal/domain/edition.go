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
	GoogleBooksID        string     `json:"google_books_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// EditionDeduplicationKey computes a unique clustering key for an edition,
// prioritizing clean ISBN-13, ISBN-10, ASIN, or OpenLibrary ID.
func EditionDeduplicationKey(ed Edition) string {
	if ed.ISBN13 != nil && strings.TrimSpace(*ed.ISBN13) != "" {
		return "isbn13:" + strings.TrimSpace(*ed.ISBN13)
	}
	if ed.ISBN10 != nil && strings.TrimSpace(*ed.ISBN10) != "" {
		return "isbn10:" + strings.TrimSpace(*ed.ISBN10)
	}
	if ed.ASIN != nil && strings.TrimSpace(*ed.ASIN) != "" {
		return "asin:" + strings.TrimSpace(*ed.ASIN)
	}
	if strings.TrimSpace(ed.OpenLibraryEditionID) != "" {
		return "olid:" + strings.TrimSpace(ed.OpenLibraryEditionID)
	}
	return "title_pub_fmt:" + strings.ToLower(strings.TrimSpace(ed.Title)) + "|" + string(ed.Format) + "|" + strings.ToLower(strings.TrimSpace(ed.Publisher))
}

// MergeEditions combines two edition records with the same identifier,
// keeping the highest quality metadata (known format, verified page count, publisher, cover).
func MergeEditions(primary, secondary Edition) Edition {
	merged := primary

	// Keep primary ID if present, else secondary
	if merged.ID == "" && secondary.ID != "" {
		merged.ID = secondary.ID
	}
	if merged.WorkID == "" && secondary.WorkID != "" {
		merged.WorkID = secondary.WorkID
	}
	// Prefer known format over UNKNOWN
	if (merged.Format == "" || merged.Format == FormatUnknown) && secondary.Format != "" && secondary.Format != FormatUnknown {
		merged.Format = secondary.Format
	}
	// Prefer non-nil page count
	if merged.PageCount == nil && secondary.PageCount != nil {
		merged.PageCount = secondary.PageCount
	}
	// Prefer non-empty publisher
	if strings.TrimSpace(merged.Publisher) == "" && strings.TrimSpace(secondary.Publisher) != "" {
		merged.Publisher = secondary.Publisher
	}
	// Prefer non-nil publication year
	if merged.PublicationYear == nil && secondary.PublicationYear != nil {
		merged.PublicationYear = secondary.PublicationYear
	}
	// Prefer non-empty publication date
	if strings.TrimSpace(merged.PublicationDate) == "" && strings.TrimSpace(secondary.PublicationDate) != "" {
		merged.PublicationDate = secondary.PublicationDate
	}
	// Prefer non-empty cover URL
	if strings.TrimSpace(merged.CoverURL) == "" && strings.TrimSpace(secondary.CoverURL) != "" {
		merged.CoverURL = secondary.CoverURL
	}
	// Prefer non-empty language
	if strings.TrimSpace(merged.Language) == "" && strings.TrimSpace(secondary.Language) != "" {
		merged.Language = secondary.Language
	}
	// Prefer non-empty description
	if strings.TrimSpace(merged.Description) == "" && strings.TrimSpace(secondary.Description) != "" {
		merged.Description = secondary.Description
	}
	// Preserve ISBNs
	if merged.ISBN10 == nil && secondary.ISBN10 != nil {
		merged.ISBN10 = secondary.ISBN10
	}
	if merged.ISBN13 == nil && secondary.ISBN13 != nil {
		merged.ISBN13 = secondary.ISBN13
	}
	if merged.ASIN == nil && secondary.ASIN != nil {
		merged.ASIN = secondary.ASIN
	}
	if merged.OpenLibraryEditionID == "" && secondary.OpenLibraryEditionID != "" {
		merged.OpenLibraryEditionID = secondary.OpenLibraryEditionID
	}

	return merged
}

// DeduplicateEditions groups editions by ISBN/identifier and merges duplicates,
// preserving original insertion order while eliminating redundant cards.
func DeduplicateEditions(editions []Edition) []Edition {
	if len(editions) <= 1 {
		return editions
	}

	seen := make(map[string]int) // key -> index in unique slice
	unique := make([]Edition, 0, len(editions))

	for _, ed := range editions {
		key := EditionDeduplicationKey(ed)
		if idx, exists := seen[key]; exists {
			// Merge into existing edition
			unique[idx] = MergeEditions(unique[idx], ed)
		} else {
			seen[key] = len(unique)
			unique = append(unique, ed)
		}
	}

	return unique
}

