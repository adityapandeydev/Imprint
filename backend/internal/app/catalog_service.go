package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// CatalogService orchestrates book discovery, metadata inspection, and lazy catalog persistence.
type CatalogService struct {
	provider        domain.BookProvider
	workRepo        domain.WorkRepository
	editionRepo     domain.EditionRepository
	searchCacheRepo domain.SearchCacheRepository
	searchCache     sync.Map
}

// NewCatalogService initializes a CatalogService with optional persistent search cache.
func NewCatalogService(
	provider domain.BookProvider,
	workRepo domain.WorkRepository,
	editionRepo domain.EditionRepository,
	searchCacheRepo ...domain.SearchCacheRepository,
) *CatalogService {
	var cacheRepo domain.SearchCacheRepository
	if len(searchCacheRepo) > 0 {
		cacheRepo = searchCacheRepo[0]
	}
	return &CatalogService{
		provider:        provider,
		workRepo:        workRepo,
		editionRepo:     editionRepo,
		searchCacheRepo: cacheRepo,
	}
}

// Search queries both local catalog and the external metadata provider,
// returning normalized, deduplicated works with fast in-memory caching
// and persistent 7-day database search query caching.
func (s *CatalogService) Search(ctx context.Context, query string, limit int) ([]domain.Work, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return []domain.Work{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	cacheKey := fmt.Sprintf("%s:%d", strings.ToLower(trimmed), limit)

	// 1. ALWAYS check local catalog first — books the user has already saved
	// must always appear in results regardless of cache or provider state.
	localWorks, _ := s.workRepo.SearchLocalWorks(ctx, trimmed, limit)

	// 2. Check in-memory cache for provider results
	if cached, ok := s.searchCache.Load(cacheKey); ok {
		return s.mergeLocalAndProvider(localWorks, cached.([]domain.Work), trimmed, limit), nil
	}

	// 3. Check persistent database cache (Neon PostgreSQL search_queries table)
	if s.searchCacheRepo != nil {
		if entry, err := s.searchCacheRepo.GetCachedQuery(ctx, trimmed); err == nil && entry != nil && len(entry.Results) > 0 {
			s.searchCache.Store(cacheKey, entry.Results)
			return s.mergeLocalAndProvider(localWorks, entry.Results, trimmed, limit), nil
		}
	}

	// 4. Query external provider (cache miss)
	providerWorks, err := s.provider.Search(ctx, domain.ProviderSearchParams{
		Query: trimmed,
		Limit: limit,
	})
	if err != nil && len(localWorks) == 0 {
		return nil, fmt.Errorf("searching book provider: %w", err)
	}

	// 4. Merge and deduplicate by OpenLibraryWorkID, GoogleBooksID, or Title
	seen := make(map[string]bool)
	var combined []domain.Work

	markSeen := func(w domain.Work) {
		if w.OpenLibraryWorkID != "" {
			seen["ol:"+w.OpenLibraryWorkID] = true
		}
		if w.GoogleBooksID != "" {
			seen["gb:"+w.GoogleBooksID] = true
		}
		seen["t:"+strings.ToLower(w.Title)] = true
	}
	isSeen := func(w domain.Work) bool {
		if w.OpenLibraryWorkID != "" && seen["ol:"+w.OpenLibraryWorkID] {
			return true
		}
		if w.GoogleBooksID != "" && seen["gb:"+w.GoogleBooksID] {
			return true
		}
		return seen["t:"+strings.ToLower(w.Title)]
	}

	for _, w := range localWorks {
		if !isSeen(w) {
			markSeen(w)
			combined = append(combined, w)
		}
	}

	for _, w := range providerWorks {
		if !isSeen(w) {
			markSeen(w)
			combined = append(combined, w)
		}
	}

	// 5. Apply Exact-Match #1 dynamic ranker across all candidate works
	combined = domain.RankWorks(combined, trimmed)
	if len(combined) > limit {
		combined = combined[:limit]
	}

	// In-memory cache stores the provider results so local DB results are
	// always freshly merged on subsequent cache hits.
	s.searchCache.Store(cacheKey, providerWorks)

	// 6. Asynchronously persist discovered works and query cache in background
	// This ensures the HTTP search response is returned immediately to the reader in ~350ms
	// without waiting on 60-80 remote database round-trips!
	worksToSave := make([]domain.Work, len(combined))
	copy(worksToSave, combined)

	go func(works []domain.Work, q string) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		for i := range works {
			_ = s.workRepo.SaveWork(bgCtx, &works[i])
		}

		if s.searchCacheRepo != nil && len(works) > 0 {
			_ = s.searchCacheRepo.SaveCachedQuery(bgCtx, &domain.SearchCacheEntry{
				QueryText:   q,
				Results:     works,
				ResultCount: len(works),
			})
		}
	}(worksToSave, trimmed)

	return combined, nil
}

// mergeLocalAndProvider merges locally saved books with provider/cached results,
// deduplicating and ranking them. This ensures books already in the user's local
// catalog always appear in search results even when served from cache.
func (s *CatalogService) mergeLocalAndProvider(local, provider []domain.Work, query string, limit int) []domain.Work {
	seen := make(map[string]bool)
	var combined []domain.Work

	markSeen := func(w domain.Work) {
		if w.OpenLibraryWorkID != "" {
			seen["ol:"+w.OpenLibraryWorkID] = true
		}
		if w.GoogleBooksID != "" {
			seen["gb:"+w.GoogleBooksID] = true
		}
		seen["t:"+strings.ToLower(w.Title)] = true
	}
	isSeen := func(w domain.Work) bool {
		if w.OpenLibraryWorkID != "" && seen["ol:"+w.OpenLibraryWorkID] {
			return true
		}
		if w.GoogleBooksID != "" && seen["gb:"+w.GoogleBooksID] {
			return true
		}
		return seen["t:"+strings.ToLower(w.Title)]
	}

	for _, w := range local {
		if !isSeen(w) {
			markSeen(w)
			combined = append(combined, w)
		}
	}
	for _, w := range provider {
		if !isSeen(w) {
			markSeen(w)
			combined = append(combined, w)
		}
	}

	combined = domain.RankWorks(combined, query)
	if len(combined) > limit {
		combined = combined[:limit]
	}
	return combined
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

	// 1. Try local lookup by UUID, Open Library ID, or Google Books ID
	if domain.IsUUID(trimmed) {
		work, err = s.workRepo.GetWorkByID(ctx, trimmed)
	} else if strings.HasPrefix(trimmed, "OL") || strings.Contains(trimmed, "/works/OL") {
		work, err = s.workRepo.GetWorkByOpenLibraryID(ctx, trimmed)
	} else {
		work, err = s.workRepo.GetWorkByGoogleBooksID(ctx, trimmed)
		if err != nil || work == nil {
			if w, wErr := s.workRepo.GetWorkByID(ctx, trimmed); wErr == nil && w != nil {
				work = w
				err = nil
			}
		}
	}

	// 3. Fall back to external provider if not found locally
	if err != nil || work == nil {
		work, err = s.provider.GetWork(ctx, trimmed)
		if err != nil {
			return nil, nil, fmt.Errorf("fetching work from provider: %w", err)
		}
		if work == nil {
			return nil, nil, domain.ErrWorkNotFound
		}
		// Lazily persist the Work to PostgreSQL
		if saveErr := s.workRepo.SaveWork(ctx, work); saveErr != nil {
			return nil, nil, fmt.Errorf("persisting work to database: %w", saveErr)
		}
	}

	// 4. Retrieve editions from local storage
	if work != nil && work.ID != "" {
		editions, _ = s.editionRepo.GetEditionsByWorkID(ctx, work.ID)
	}

	// 5. If no editions are stored locally yet, fetch from provider and cache them
	if len(editions) == 0 {
		lookupKey := work.OpenLibraryWorkID
		if lookupKey == "" {
			lookupKey = work.GoogleBooksID
		}
		if lookupKey == "" {
			lookupKey = trimmed
		}
		providerEditions, pErr := s.provider.GetEditionsForWork(ctx, lookupKey, 20)
		if (pErr != nil || len(providerEditions) == 0) && work.Title != "" {
			// Secondary fallback: query editions by title
			providerEditions, _ = s.provider.GetEditionsForWork(ctx, work.Title, 20)
		}
		if len(providerEditions) > 0 {
			providerEditions = domain.DeduplicateEditions(providerEditions)
			for i := range providerEditions {
				ed := &providerEditions[i]
				ed.WorkID = work.ID
				if sErr := s.editionRepo.SaveEdition(ctx, ed); sErr == nil {
					editions = append(editions, *ed)
				}
			}
		}
	}

	// 6. Ensure returned editions are clean, merged, and deduplicated
	editions = domain.DeduplicateEditions(editions)

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
