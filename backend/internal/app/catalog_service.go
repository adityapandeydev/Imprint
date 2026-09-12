package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// CatalogService orchestrates book discovery, metadata inspection, and lazy catalog persistence.
type CatalogService struct {
	provider    domain.BookProvider
	workRepo    domain.WorkRepository
	editionRepo domain.EditionRepository
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
// returning normalized, deduplicated works without polluting PostgreSQL with transient results.
func (s *CatalogService) Search(ctx context.Context, query string, limit int) ([]domain.Work, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []domain.Work{}, nil
	}
	if limit <= 0 {
		limit = 20
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

	return combined, nil
}

// GetWork fetches a Work and its Editions.
// If the work is not yet stored locally, it is fetched from the external provider
// and lazily persisted into PostgreSQL (per ADR 0002).
func (s *CatalogService) GetWork(ctx context.Context, idOrOLID string) (*domain.Work, []domain.Edition, error) {
	trimmed := strings.TrimSpace(idOrOLID)
	if trimmed == "" {
		return nil, nil, domain.ErrWorkNotFound
	}

	// 1. Try local lookup by UUID
	work, err := s.workRepo.GetWorkByID(ctx, trimmed)
	if err == nil {
		editions, _ := s.editionRepo.GetEditionsByWorkID(ctx, work.ID)
		return work, editions, nil
	}

	// 2. Try local lookup by Open Library ID
	work, err = s.workRepo.GetWorkByOpenLibraryID(ctx, trimmed)
	if err == nil {
		editions, _ := s.editionRepo.GetEditionsByWorkID(ctx, work.ID)
		return work, editions, nil
	}

	// 3. Fall back to external provider
	work, err = s.provider.GetWork(ctx, trimmed)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching work from provider: %w", err)
	}

	// Lazily persist the Work to PostgreSQL
	if err := s.workRepo.SaveWork(ctx, work); err != nil {
		return nil, nil, fmt.Errorf("persisting work to database: %w", err)
	}

	// Fetch published editions from provider and persist them
	providerEditions, err := s.provider.GetEditionsForWork(ctx, trimmed, 20)
	var savedEditions []domain.Edition
	if err == nil {
		for i := range providerEditions {
			ed := &providerEditions[i]
			ed.WorkID = work.ID
			if err := s.editionRepo.SaveEdition(ctx, ed); err == nil {
				savedEditions = append(savedEditions, *ed)
			}
		}
	}

	return work, savedEditions, nil
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
