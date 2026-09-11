package openlibrary

import (
	"fmt"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// cleanOLKey strips any leading path (e.g. "/works/OL27479W" -> "OL27479W").
func cleanOLKey(raw string) string {
	parts := strings.Split(strings.Trim(raw, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return raw
}

// parseDescription extracts description text whether Open Library returns a plain string
// or an object with {"type": "/type/text", "value": "..."}.
func parseDescription(raw any) string {
	if raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v)
	case map[string]any:
		if val, ok := v["value"].(string); ok {
			return strings.TrimSpace(val)
		}
	}
	return ""
}

// buildCoverURL constructs the full URL for an Open Library cover ID.
func buildCoverURL(coverID int, coverBaseURL string) string {
	if coverID <= 0 {
		return ""
	}
	return fmt.Sprintf("%s/b/id/%d-L.jpg", strings.TrimRight(coverBaseURL, "/"), coverID)
}

// mapSearchDocToWork converts a search result document into a domain.Work entity.
func mapSearchDocToWork(doc searchDocDTO, coverBaseURL string) domain.Work {
	var authors []domain.Author
	for _, name := range doc.AuthorName {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			authors = append(authors, domain.Author{
				Name: trimmed,
			})
		}
	}

	var originalYear *int
	if doc.FirstPublishYear > 0 {
		yr := doc.FirstPublishYear
		originalYear = &yr
	}

	var coverURL string
	if doc.CoverI > 0 {
		coverURL = buildCoverURL(doc.CoverI, coverBaseURL)
	}

	return domain.Work{
		Title:             strings.TrimSpace(doc.Title),
		OriginalYear:      originalYear,
		CoverURL:          coverURL,
		OpenLibraryWorkID: cleanOLKey(doc.Key),
		SubjectTags:       doc.Subject,
		Authors:           authors,
	}
}

// mapWorkDTOToWork converts a /works/{id}.json response into a domain.Work.
func mapWorkDTOToWork(dto workResponseDTO, coverBaseURL string) domain.Work {
	var coverURL string
	if len(dto.Covers) > 0 && dto.Covers[0] > 0 {
		coverURL = buildCoverURL(dto.Covers[0], coverBaseURL)
	}

	return domain.Work{
		Title:             strings.TrimSpace(dto.Title),
		Description:       parseDescription(dto.Description),
		CoverURL:          coverURL,
		OpenLibraryWorkID: cleanOLKey(dto.Key),
		SubjectTags:       dto.Subjects,
	}
}

// mapEditionDTOToEdition converts an edition DTO into a domain.Edition.
func mapEditionDTOToEdition(dto editionDTO, workID string, coverBaseURL string) domain.Edition {
	publisher := ""
	if len(dto.Publishers) > 0 {
		publisher = strings.TrimSpace(dto.Publishers[0])
	}

	var isbn10 *string
	if len(dto.ISBN10) > 0 {
		cleaned := domain.CleanIdentifier(dto.ISBN10[0])
		if domain.ValidateISBN10(cleaned) {
			isbn10 = &cleaned
		}
	}

	var isbn13 *string
	if len(dto.ISBN13) > 0 {
		cleaned := domain.CleanIdentifier(dto.ISBN13[0])
		if domain.ValidateISBN13(cleaned) {
			isbn13 = &cleaned
		}
	}

	// If no ISBN-13 but ISBN-10 exists, compute ISBN-13
	if isbn13 == nil && isbn10 != nil {
		if converted, err := domain.ISBN10To13(*isbn10); err == nil {
			isbn13 = &converted
		}
	}

	var pageCount *int
	if dto.NumberOfPages > 0 {
		pages := dto.NumberOfPages
		pageCount = &pages
	}

	var coverURL string
	if len(dto.Covers) > 0 && dto.Covers[0] > 0 {
		coverURL = buildCoverURL(dto.Covers[0], coverBaseURL)
	} else if isbn13 != nil {
		coverURL = fmt.Sprintf("%s/b/isbn/%s-L.jpg", strings.TrimRight(coverBaseURL, "/"), *isbn13)
	}

	return domain.Edition{
		WorkID:               workID,
		Title:                strings.TrimSpace(dto.Title),
		Publisher:            publisher,
		PublicationDate:      strings.TrimSpace(dto.PublishDate),
		PageCount:            pageCount,
		Format:               domain.ParseBookFormat(dto.PhysicalFormat),
		ISBN10:               isbn10,
		ISBN13:               isbn13,
		CoverURL:             coverURL,
		Description:          parseDescription(dto.Description),
		OpenLibraryEditionID: cleanOLKey(dto.Key),
	}
}
