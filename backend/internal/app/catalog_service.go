package app

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// CatalogService orchestrates book discovery, metadata inspection, and lazy catalog persistence.
type CatalogService struct {
	provider    domain.BookProvider
	workRepo    domain.WorkRepository
	editionRepo domain.EditionRepository
	searchCache sync.Map
}

// NewCatalogService initializes a CatalogService.
func NewCatalogService(
	provider domain.BookProvider,
	workRepo domain.WorkRepository,
	editionRepo domain.EditionRepository,
) *CatalogService {
	return &CatalogService{
		provider:    provider,
		workRepo:    workRepo,
		editionRepo: editionRepo,
	}
}

// Search queries both local catalog and the external metadata provider,
// returning normalized, deduplicated works with fast in-memory caching.
func (s *CatalogService) Search(ctx context.Context, query string, limit int) ([]domain.Work, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []domain.Work{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	cacheKey := fmt.Sprintf("%s:%d", strings.ToLower(trimmed), limit)
	if cached, ok := s.searchCache.Load(cacheKey); ok {
		return cached.([]domain.Work), nil
	}

	// 1. Check local catalog first
	localWorks, _ := s.workRepo.SearchLocalWorks(ctx, trimmed, limit)

	// 2. Query external provider
	providerWorks, err := s.provider.Search(ctx, domain.ProviderSearchParams{
		Query: trimmed,
		Limit: limit,
	})
	if err != nil && len(localWorks) == 0 {
		return nil, fmt.Errorf("searching book provider: %w", err)
	}

	// 3. Merge and deduplicate by OpenLibraryWorkID or Title
	seen := make(map[string]bool)
	var combined []domain.Work

	for _, w := range localWorks {
		key := strings.ToLower(w.Title)
		if w.OpenLibraryWorkID != "" {
			key = w.OpenLibraryWorkID
		}
		seen[key] = true
		combined = append(combined, w)
	}

	for _, w := range providerWorks {
		key := strings.ToLower(w.Title)
		if w.OpenLibraryWorkID != "" {
			key = w.OpenLibraryWorkID
		}
		if !seen[key] {
			seen[key] = true
			combined = append(combined, w)
		}
		if len(combined) >= limit {
			break
		}
	}

	// Cache discovered works locally so later GetWork and AddToWishlist have immediate access
	for i := range combined {
		_ = s.workRepo.SaveWork(ctx, &combined[i])
	}
	s.searchCache.Store(cacheKey, combined)

	return combined, nil
}

// GetWork fetches a Work and its Editions.
// If the work is not yet stored locally, it is fetched from the external provider
// and lazily persisted into PostgreSQL (per ADR 0002).
// If editions are not yet cached locally, they are resolved from the provider and saved.
func (s *CatalogService) GetWork(ctx context.Context, idOrOLID string) (*domain.Work, []domain.Edition, error) {
	trimmed := strings.TrimSpace(idOrOLID)
	if trimmed == "" {
		return nil, nil, domain.ErrWorkNotFound
	}

	var work *domain.Work
	var editions []domain.Edition
	var err error

	// 1. Try local lookup by UUID
	work, err = s.workRepo.GetWorkByID(ctx, trimmed)
	if err != nil {
		// 2. Try local lookup by Open Library ID
		work, err = s.workRepo.GetWorkByOpenLibraryID(ctx, trimmed)
	}

	// 3. Fall back to external provider if not found locally
	if err != nil || work == nil {
		work, err = s.provider.GetWork(ctx, trimmed)
		if err != nil {
			return nil, nil, fmt.Errorf("fetching work from provider: %w", err)
		}
		// Lazily persist the Work to PostgreSQL
		if saveErr := s.workRepo.SaveWork(ctx, work); saveErr != nil {
			return nil, nil, fmt.Errorf("persisting work to database: %w", saveErr)
		}
	}

	// 4. Retrieve editions from local storage
	if work.ID != "" {
		editions, _ = s.editionRepo.GetEditionsByWorkID(ctx, work.ID)
	}

	// 5. If no editions are stored locally yet, fetch from provider and cache them
	if len(editions) == 0 {
		lookupKey := work.OpenLibraryWorkID
		if lookupKey == "" {
			lookupKey = trimmed
		}
		providerEditions, pErr := s.provider.GetEditionsForWork(ctx, lookupKey, 20)
		if (pErr != nil || len(providerEditions) == 0) && work.Title != "" {
			// Secondary fallback: query editions by title
			providerEditions, _ = s.provider.GetEditionsForWork(ctx, work.Title, 20)
		}
		if len(providerEditions) > 0 {
			for i := range providerEditions {
				ed := &providerEditions[i]
				ed.WorkID = work.ID
				if sErr := s.editionRepo.SaveEdition(ctx, ed); sErr == nil {
					editions = append(editions, *ed)
				}
			}
		}
	}

	return work, editions, nil
}

// GetEditionByISBN locates an edition by ISBN-10, ISBN-13, or ASIN.
// Lazily persists both Work and Edition if not yet saved locally.
func (s *CatalogService) GetEditionByISBN(ctx context.Context, rawISBN string) (*domain.Edition, *domain.Work, error) {
	cleaned := domain.CleanIdentifier(rawISBN)
	if cleaned == "" {
		return nil, nil, domain.ErrInvalidISBN
	}

	// 1. Check local database
	edition, err := s.editionRepo.GetEditionByISBN(ctx, cleaned)
	if err == nil {
		work, err := s.workRepo.GetWorkByID(ctx, edition.WorkID)
		if err == nil {
			return edition, work, nil
		}
	}

	// 2. Query external provider
	edition, work, err := s.provider.GetEditionByISBN(ctx, cleaned)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching edition by isbn from provider: %w", err)
	}

	// 3. Persist work if available
	if work != nil {
		if err := s.workRepo.SaveWork(ctx, work); err == nil {
			edition.WorkID = work.ID
		}
	}

	// 4. Persist edition
	if err := s.editionRepo.SaveEdition(ctx, edition); err != nil {
		return nil, nil, fmt.Errorf("persisting edition: %w", err)
	}

	return edition, work, nil
}
