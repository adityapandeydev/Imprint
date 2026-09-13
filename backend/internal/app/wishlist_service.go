package app

import (
	"context"
	"fmt"
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
	Notes     *string               `json:"notes,omitempty"`
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
		}
		if items[i].Edition == nil && items[i].EditionID != nil && *items[i].EditionID != "" {
			if ed, err := s.editionRepo.GetEditionByID(ctx, *items[i].EditionID); err == nil {
				items[i].Edition = ed
			}
		}
	}

	return items, nil
}

// AddToWishlist adds a book to a user's collection.
// If the work is not yet stored locally, it ensures the work is lazily persisted first.
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

	item := domain.WishlistItem{
		UserID:    req.UserID,
		WorkID:    work.ID,
		Work:      work,
		EditionID: preferredEditionID,
		Edition:   preferredEdition,
		Status:    status,
		Priority:  priority,
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

// UpdateWishlistItem updates status, priority, rating, or notes for a wishlist entry.
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

	if err := item.Validate(); err != nil {
		return nil, err
	}

	if err := s.wishlistRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("updating wishlist item: %w", err)
	}

	res, err := s.wishlistRepo.GetByID(ctx, item.ID)
	if err == nil {
		if res.Work == nil {
			res.Work = item.Work
		}
		if res.Edition == nil {
			res.Edition = item.Edition
		}
		return res, nil
	}

	return item, nil
}

// RemoveFromWishlist removes an item from a user's collection.
func (s *WishlistService) RemoveFromWishlist(ctx context.Context, itemID, userID string) error {
	return s.wishlistRepo.Delete(ctx, itemID, userID)
}
