package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// AddToWishlistRequest encapsulates input for adding a work to a user's collection.
type AddToWishlistRequest struct {
	UserID       string               `json:"user_id"`
	WorkID       string               `json:"work_id"` // Local UUID or Open Library Work ID
	EditionID    *string              `json:"edition_id,omitempty"`
	Status       domain.ReadingStatus `json:"status"`
	Priority     int                  `json:"priority"`
	Rating       *int                 `json:"rating,omitempty"`
	Tags         []string             `json:"tags,omitempty"`
	Notes        string               `json:"notes,omitempty"`
	Title        string               `json:"title,omitempty"`
	Author       string               `json:"author,omitempty"`
	CoverURL     string               `json:"cover_url,omitempty"`
	OriginalYear *int                 `json:"original_year,omitempty"`
}

// UpdateWishlistRequest encapsulates input for modifying an existing wishlist item.
type UpdateWishlistRequest struct {
	ID        string                `json:"id"`
	UserID    string                `json:"user_id"`
	EditionID *string               `json:"edition_id,omitempty"`
	Status    *domain.ReadingStatus `json:"status,omitempty"`
	Priority  *int                  `json:"priority,omitempty"`
	Rating    *int                  `json:"rating,omitempty"`
	Tags      *[]string             `json:"tags,omitempty"`
	Notes     *string               `json:"notes,omitempty"`
}

// GoodreadsImportSummary aggregates results of a batch CSV import.
type GoodreadsImportSummary struct {
	TotalRows     int `json:"total_rows"`
	ImportedCount int `json:"imported_count"`
	SkippedCount  int `json:"skipped_count"`
	FailedCount   int `json:"failed_count"`
}

// WishlistService orchestrates personal collection management and reading lifecycle tracking.
type WishlistService struct {
	wishlistRepo   domain.WishlistRepository
	catalogService *CatalogService
	workRepo       domain.WorkRepository
	editionRepo    domain.EditionRepository
}

// NewWishlistService initializes a WishlistService.
func NewWishlistService(
	wishlistRepo domain.WishlistRepository,
	catalogService *CatalogService,
	workRepo domain.WorkRepository,
	editionRepo domain.EditionRepository,
) *WishlistService {
	return &WishlistService{
		wishlistRepo:   wishlistRepo,
		catalogService: catalogService,
		workRepo:       workRepo,
		editionRepo:    editionRepo,
	}
}

// ListUserWishlist retrieves a user's reading collection, optionally filtered by status.
func (s *WishlistService) ListUserWishlist(
	ctx context.Context,
	userID string,
	statusFilter *domain.ReadingStatus,
) ([]domain.WishlistItem, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrUserNotFound
	}
	items, err := s.wishlistRepo.ListByUser(ctx, userID, statusFilter)
	if err != nil {
		return nil, err
	}

	for i := range items {
		if items[i].Work == nil && items[i].WorkID != "" {
			if w, err := s.workRepo.GetWorkByID(ctx, items[i].WorkID); err == nil {
				items[i].Work = w
			}
		} else if items[i].Work != nil && len(items[i].Work.Authors) == 0 && items[i].WorkID != "" {
			if w, err := s.workRepo.GetWorkByID(ctx, items[i].WorkID); err == nil && len(w.Authors) > 0 {
				items[i].Work.Authors = w.Authors
			}
		}
		if items[i].Edition == nil && items[i].EditionID != nil && *items[i].EditionID != "" {
			if ed, err := s.editionRepo.GetEditionByID(ctx, *items[i].EditionID); err == nil {
				items[i].Edition = ed
			}
		}
	}

	return items, nil
}

// GetUserTags returns all custom tags created by the reader with book counts.
func (s *WishlistService) GetUserTags(ctx context.Context, userID string) ([]domain.TagCount, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrUserNotFound
	}
	return s.wishlistRepo.GetUserTags(ctx, userID)
}

// AddToWishlist adds a book to a user's collection.
func (s *WishlistService) AddToWishlist(ctx context.Context, req AddToWishlistRequest) (*domain.WishlistItem, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return nil, domain.ErrUserNotFound
	}
	if strings.TrimSpace(req.WorkID) == "" {
		return nil, domain.ErrWorkNotFound
	}

	// 1. Ensure Work exists locally (or resolve from provider lazily)
	var work *domain.Work
	var editions []domain.Edition

	// First try local lookups
	if w, err := s.workRepo.GetWorkByID(ctx, req.WorkID); err == nil {
		work = w
	} else if w, err := s.workRepo.GetWorkByOpenLibraryID(ctx, req.WorkID); err == nil {
		work = w
	}

	// If not found locally, but client provided metadata, save locally immediately
	if work == nil && req.Title != "" {
		newWork := &domain.Work{
			Title:             req.Title,
			CoverURL:          req.CoverURL,
			OriginalYear:      req.OriginalYear,
			OpenLibraryWorkID: req.WorkID,
		}
		if req.Author != "" {
			newWork.Authors = []domain.Author{{Name: req.Author}}
		}
		_ = s.workRepo.SaveWork(ctx, newWork)
		work = newWork
	}

	// If still not found, resolve from provider
	if work == nil {
		var err error
		work, editions, err = s.catalogService.GetWork(ctx, req.WorkID)
		if err != nil {
			return nil, fmt.Errorf("resolving work: %w", err)
		}
	}

	if work != nil && len(work.Authors) == 0 && req.Author != "" {
		work.Authors = []domain.Author{{Name: req.Author}}
		_ = s.workRepo.SaveWork(ctx, work)
	}

	// 2. Check if already in user's collection
	existing, _ := s.wishlistRepo.GetByUserAndWork(ctx, req.UserID, work.ID)
	if existing != nil {
		return nil, domain.ErrDuplicateWishlistItem
	}

	// 3. Resolve preferred edition if not provided
	var preferredEditionID *string
	var preferredEdition *domain.Edition
	if req.EditionID != nil && *req.EditionID != "" {
		if ed, err := s.editionRepo.GetEditionByID(ctx, *req.EditionID); err == nil {
			preferredEditionID = req.EditionID
			preferredEdition = ed
		}
	} else if len(editions) > 0 {
		preferredEditionID = &editions[0].ID
		preferredEdition = &editions[0]
	}

	// Default status and priority
	status := req.Status
	if status == "" {
		status = domain.StatusWantToRead
	}
	priority := req.Priority
	if priority < 1 || priority > 5 {
		priority = 3
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	item := domain.WishlistItem{
		UserID:    req.UserID,
		WorkID:    work.ID,
		Work:      work,
		EditionID: preferredEditionID,
		Edition:   preferredEdition,
		Status:    status,
		Priority:  priority,
		Rating:    req.Rating,
		Tags:      tags,
		Notes:     strings.TrimSpace(req.Notes),
	}

	if status == domain.StatusCurrentlyReading {
		now := time.Now()
		item.StartedAt = &now
	} else if status == domain.StatusFinished {
		now := time.Now()
		item.FinishedAt = &now
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	if err := s.wishlistRepo.Save(ctx, &item); err != nil {
		return nil, fmt.Errorf("saving wishlist item: %w", err)
	}

	fetched, err := s.wishlistRepo.GetByID(ctx, item.ID)
	if err == nil {
		if fetched.Work == nil {
			fetched.Work = work
		}
		if fetched.Edition == nil {
			fetched.Edition = preferredEdition
		}
		return fetched, nil
	}

	return &item, nil
}

// UpdateWishlistItem updates status, priority, rating, notes, or tags for a wishlist entry.
func (s *WishlistService) UpdateWishlistItem(ctx context.Context, req UpdateWishlistRequest) (*domain.WishlistItem, error) {
	item, err := s.wishlistRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Security check: ensure item belongs to user
	if item.UserID != req.UserID {
		return nil, domain.ErrNotFound
	}

	if req.EditionID != nil {
		if *req.EditionID == "" {
			item.EditionID = nil
			item.Edition = nil
		} else {
			item.EditionID = req.EditionID
			if ed, err := s.editionRepo.GetEditionByID(ctx, *req.EditionID); err == nil {
				item.Edition = ed
			}
		}
	}

	if req.Status != nil {
		validStatus, err := domain.ValidateReadingStatus(string(*req.Status))
		if err != nil {
			return nil, err
		}
		// Track reading lifecycle timestamps automatically
		if validStatus == domain.StatusCurrentlyReading && item.StartedAt == nil {
			now := time.Now()
			item.StartedAt = &now
		} else if validStatus == domain.StatusFinished && item.FinishedAt == nil {
			now := time.Now()
			item.FinishedAt = &now
		}
		item.Status = validStatus
	}

	if req.Priority != nil {
		if *req.Priority < 1 || *req.Priority > 5 {
			return nil, domain.ErrInvalidPriority
		}
		item.Priority = *req.Priority
	}

	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, domain.ErrInvalidRating
		}
		item.Rating = req.Rating
	}

	if req.Notes != nil {
		item.Notes = strings.TrimSpace(*req.Notes)
	}

	if req.Tags != nil {
		cleanedTags := make([]string, 0, len(*req.Tags))
		seen := make(map[string]bool)
		for _, t := range *req.Tags {
			clean := strings.TrimSpace(t)
			if clean != "" && !seen[strings.ToLower(clean)] {
				seen[strings.ToLower(clean)] = true
				cleanedTags = append(cleanedTags, clean)
			}
		}
		item.Tags = cleanedTags
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	if err := s.wishlistRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("updating wishlist item: %w", err)
	}

	return item, nil
}

// RemoveFromWishlist removes an item from a user's collection.
func (s *WishlistService) RemoveFromWishlist(ctx context.Context, itemID, userID string) error {
	return s.wishlistRepo.Delete(ctx, itemID, userID)
}

// cleanGoodreadsField strips spreadsheet formula wrappers like ="0385537859"
func cleanGoodreadsField(val string) string {
	s := strings.TrimSpace(val)
	s = strings.TrimPrefix(s, "=")
	s = strings.Trim(s, "\"")
	return strings.TrimSpace(s)
}

// parseDate attempts multiple date formats common in Goodreads CSV exports.
func parseDate(val string) *time.Time {
	clean := cleanGoodreadsField(val)
	if clean == "" {
		return nil
	}
	formats := []string{
		"2006/01/02",
		"2006-01-02",
		"01/02/2006",
		"1/2/2006",
		"02/01/2006",
		"2006/01",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, clean); err == nil {
			return &t
		}
	}
	return nil
}

// ImportGoodreadsCSV parses an uploaded Goodreads library export CSV,
// normalizes ISBNs, maps exclusive shelves to ReadingStatus, creates custom tags, and saves books.
func (s *WishlistService) ImportGoodreadsCSV(ctx context.Context, userID string, r io.Reader) (*GoodreadsImportSummary, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrUserNotFound
	}

	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // Allow variable field count gracefully

	headerRow, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	colMap := make(map[string]int)
	for i, h := range headerRow {
		colMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	getCol := func(row []string, name string) string {
		idx, ok := colMap[name]
		if !ok || idx >= len(row) {
			return ""
		}
		return row[idx]
	}

	summary := &GoodreadsImportSummary{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			summary.FailedCount++
			continue
		}

		summary.TotalRows++

		title := cleanGoodreadsField(getCol(row, "title"))
		if title == "" {
			summary.SkippedCount++
			continue
		}

		author := cleanGoodreadsField(getCol(row, "author"))
		rawISBN13 := cleanGoodreadsField(getCol(row, "isbn13"))
		rawISBN := cleanGoodreadsField(getCol(row, "isbn"))
		exclusiveShelf := strings.ToLower(cleanGoodreadsField(getCol(row, "exclusive shelf")))
		bookshelves := cleanGoodreadsField(getCol(row, "bookshelves"))
		rawRating := cleanGoodreadsField(getCol(row, "my rating"))
		dateRead := parseDate(getCol(row, "date read"))
		dateAdded := parseDate(getCol(row, "date added"))
		notes := cleanGoodreadsField(getCol(row, "my review"))
		if notes == "" {
			notes = cleanGoodreadsField(getCol(row, "private notes"))
		}

		// 1. Determine Reading Status
		status := domain.StatusWantToRead
		if strings.Contains(exclusiveShelf, "read") && !strings.Contains(exclusiveShelf, "to-read") {
			status = domain.StatusFinished
		} else if strings.Contains(exclusiveShelf, "currently") {
			status = domain.StatusCurrentlyReading
		} else if strings.Contains(exclusiveShelf, "abandon") || strings.Contains(exclusiveShelf, "did-not-finish") {
			status = domain.StatusAbandoned
		}

		// 2. Normalize ISBN
		var isbn13 string
		if rawISBN13 != "" {
			if norm13, _, err := domain.NormalizeISBN(rawISBN13); err == nil {
				isbn13 = norm13
			}
		}
		if isbn13 == "" && rawISBN != "" {
			if norm13, _, err := domain.NormalizeISBN(rawISBN); err == nil {
				isbn13 = norm13
			}
		}

		// 3. Extract Tags
		tags := make([]string, 0)
		seenTags := make(map[string]bool)
		if bookshelves != "" {
			for _, part := range strings.Split(bookshelves, ",") {
				t := strings.TrimSpace(part)
				if t != "" && t != "to-read" && t != "currently-reading" && t != "read" {
					lower := strings.ToLower(t)
					if !seenTags[lower] {
						seenTags[lower] = true
						tags = append(tags, t)
					}
				}
			}
		}

		// 4. Rating
		var rating *int
		if rVal, err := strconv.Atoi(rawRating); err == nil && rVal >= 1 && rVal <= 5 {
			rating = &rVal
		}

		// 5. Work creation or lookup
		var work *domain.Work
		if isbn13 != "" {
			if ed, err := s.editionRepo.GetEditionByISBN(ctx, isbn13); err == nil && ed != nil {
				work, _ = s.workRepo.GetWorkByID(ctx, ed.WorkID)
			}
		}

		if work == nil {
			// Save new work entry
			newWork := &domain.Work{
				Title: title,
			}
			if author != "" {
				newWork.Authors = []domain.Author{{Name: author}}
			}
			if err := s.workRepo.SaveWork(ctx, newWork); err == nil {
				work = newWork
			} else {
				summary.FailedCount++
				continue
			}
		}

		// 6. Check duplicates
		existing, _ := s.wishlistRepo.GetByUserAndWork(ctx, userID, work.ID)
		if existing != nil {
			summary.SkippedCount++
			continue
		}

		// 7. Save WishlistItem
		item := &domain.WishlistItem{
			UserID:     userID,
			WorkID:     work.ID,
			Status:     status,
			Priority:   3,
			Rating:     rating,
			Tags:       tags,
			Notes:      notes,
			FinishedAt: dateRead,
			StartedAt:  dateAdded,
		}

		if err := s.wishlistRepo.Save(ctx, item); err != nil {
			summary.FailedCount++
		} else {
			summary.ImportedCount++
		}
	}

	return summary, nil
}

// ExportLibrary exports the user's complete collection in CSV or JSON format.
func (s *WishlistService) ExportLibrary(ctx context.Context, userID string, format string) ([]byte, string, error) {
	items, err := s.ListUserWishlist(ctx, userID, nil)
	if err != nil {
		return nil, "", err
	}

	if strings.ToLower(format) == "json" {
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return nil, "", fmt.Errorf("encoding JSON export: %w", err)
		}
		return data, "application/json", nil
	}

	// Default CSV format
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV Header
	_ = writer.Write([]string{
		"Title",
		"Author",
		"Original Year",
		"Reading Status",
		"Priority",
		"Rating",
		"Custom Tags",
		"ISBN13",
		"Format",
		"Started At",
		"Finished At",
		"Notes",
	})

	for _, item := range items {
		title := ""
		author := ""
		yearStr := ""
		if item.Work != nil {
			title = item.Work.Title
			if len(item.Work.Authors) > 0 {
				author = item.Work.Authors[0].Name
			}
			if item.Work.OriginalYear != nil {
				yearStr = strconv.Itoa(*item.Work.OriginalYear)
			}
		}

		isbn := ""
		fmtStr := ""
		if item.Edition != nil {
			if item.Edition.ISBN13 != nil {
				isbn = *item.Edition.ISBN13
			}
			fmtStr = string(item.Edition.Format)
		}

		ratingStr := ""
		if item.Rating != nil {
			ratingStr = strconv.Itoa(*item.Rating)
		}

		startedStr := ""
		if item.StartedAt != nil {
			startedStr = item.StartedAt.Format("2006-01-02")
		}

		finishedStr := ""
		if item.FinishedAt != nil {
			finishedStr = item.FinishedAt.Format("2006-01-02")
		}

		tagsStr := strings.Join(item.Tags, ", ")

		_ = writer.Write([]string{
			title,
			author,
			yearStr,
			string(item.Status),
			strconv.Itoa(item.Priority),
			ratingStr,
			tagsStr,
			isbn,
			fmtStr,
			startedStr,
			finishedStr,
			item.Notes,
		})
	}

	writer.Flush()
	return buf.Bytes(), "text/csv; charset=utf-8", nil
}
