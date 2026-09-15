package googlebooks

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

var yearRegex = regexp.MustCompile(`\b(1[0-9]{3}|20[0-9]{2})\b`)

// extractCoverURL selects the highest resolution image available and secures it with HTTPS.
func extractCoverURL(links imageLinksDTO) string {
	candidates := []string{
		links.ExtraLarge,
		links.Large,
		links.Medium,
		links.Thumbnail,
		links.SmallThumbnail,
		links.Small,
	}

	for _, raw := range candidates {
		trimmed := strings.TrimSpace(raw)
		if trimmed != "" {
			// Upgrade http:// to https://
			if strings.HasPrefix(trimmed, "http://") {
				trimmed = "https://" + strings.TrimPrefix(trimmed, "http://")
			}
			// Remove edge curling parameter for clean flat cover
			trimmed = strings.ReplaceAll(trimmed, "&edge=curl", "")
			return trimmed
		}
	}
	return ""
}

// parseYear extracts the 4-digit publication year from varied date formats.
func parseYear(dateStr string) *int {
	match := yearRegex.FindString(dateStr)
	if match != "" {
		if yr, err := strconv.Atoi(match); err == nil && yr > 0 {
			return &yr
		}
	}
	return nil
}

// extractISBNs inspects IndustryIdentifiers and extracts validated ISBN-10 and ISBN-13.
func extractISBNs(identifiers []industryIdentifierDTO) (isbn10 *string, isbn13 *string) {
	for _, id := range identifiers {
		cleaned := domain.CleanIdentifier(id.Identifier)
		switch id.Type {
		case "ISBN_10":
			if domain.ValidateISBN10(cleaned) {
				isbn10 = &cleaned
			}
		case "ISBN_13":
			if domain.ValidateISBN13(cleaned) {
				isbn13 = &cleaned
			}
		default:
			// If type was not explicitly labeled, attempt validation by length
			if len(cleaned) == 10 && domain.ValidateISBN10(cleaned) {
				isbn10 = &cleaned
			} else if len(cleaned) == 13 && domain.ValidateISBN13(cleaned) {
				isbn13 = &cleaned
			}
		}
	}
	return isbn10, isbn13
}

// mapVolumeToWork converts a Google Books volume into a domain.Work entity.
func mapVolumeToWork(v volumeResponseDTO) domain.Work {
	var authors []domain.Author
	for _, name := range v.VolumeInfo.Authors {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			authors = append(authors, domain.Author{Name: trimmed})
		}
	}

	title := strings.TrimSpace(v.VolumeInfo.Title)
	if title == "" {
		title = "Untitled Book"
	}

	return domain.Work{
		ID:            v.ID,
		GoogleBooksID: v.ID,
		Title:         title,
		Authors:       authors,
		Description:   strings.TrimSpace(v.VolumeInfo.Description),
		OriginalYear:  parseYear(v.VolumeInfo.PublishedDate),
		CoverURL:      extractCoverURL(v.VolumeInfo.ImageLinks),
		SubjectTags:   v.VolumeInfo.Categories,
	}
}

// mapVolumeToEdition converts a Google Books volume into a domain.Edition entity.
func mapVolumeToEdition(v volumeResponseDTO) domain.Edition {
	isbn10, isbn13 := extractISBNs(v.VolumeInfo.IndustryIdentifiers)

	format := domain.FormatPaperback
	if strings.EqualFold(v.VolumeInfo.PrintType, "MAGAZINE") {
		format = domain.BookFormat("MAGAZINE")
	}

	var pageCount *int
	if v.VolumeInfo.PageCount > 0 {
		pc := v.VolumeInfo.PageCount
		pageCount = &pc
	}

	lang := strings.ToLower(strings.TrimSpace(v.VolumeInfo.Language))

	return domain.Edition{
		ID:              v.ID,
		GoogleBooksID:   v.ID,
		Title:           strings.TrimSpace(v.VolumeInfo.Title),
		Format:          format,
		Publisher:       strings.TrimSpace(v.VolumeInfo.Publisher),
		PublicationYear: parseYear(v.VolumeInfo.PublishedDate),
		ISBN10:          isbn10,
		ISBN13:          isbn13,
		PageCount:       pageCount,
		Language:        lang,
		CoverURL:        extractCoverURL(v.VolumeInfo.ImageLinks),
	}
}
