package composite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// MultiProvider orchestrates multiple BookProvider implementations (e.g. Open Library + Google Books)
// to provide high availability, automated fallback, and cross-provider metadata enrichment.
type MultiProvider struct {
	primary   domain.BookProvider
	secondary domain.BookProvider
}

// NewMultiProvider creates a hybrid provider with a primary and secondary provider.
func NewMultiProvider(primary, secondary domain.BookProvider) *MultiProvider {
	return &MultiProvider{
		primary:   primary,
		secondary: secondary,
	}
}

// Name returns the composite provider identifier.
func (m *MultiProvider) Name() string {
	pName := "none"
	if m.primary != nil {
		pName = m.primary.Name()
	}
	sName := "none"
	if m.secondary != nil {
		sName = m.secondary.Name()
	}
	return fmt.Sprintf("composite(%s,%s)", pName, sName)
}

// normalizeBookKey produces a deduplication key based on title and author.
func normalizeBookKey(w domain.Work) string {
	if w.OpenLibraryWorkID != "" {
		return "ol:" + strings.ToLower(w.OpenLibraryWorkID)
	}
	title := strings.ToLower(strings.TrimSpace(w.Title))
	author := ""
	if len(w.Authors) > 0 {
		author = strings.ToLower(strings.TrimSpace(w.Authors[0].Name))
	}
	return fmt.Sprintf("%s|%s", title, author)
}

// Search queries both providers concurrently, merges results, deduplicates,
// and enriches missing covers or synopses from the complementary provider.
func (m *MultiProvider) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	if m.primary == nil && m.secondary == nil {
		return []domain.Work{}, nil
	}
	if m.primary != nil && m.secondary == nil {
		return m.primary.Search(ctx, params)
	}
	if m.primary == nil && m.secondary != nil {
		return m.secondary.Search(ctx, params)
	}

	type searchResult struct {
		works []domain.Work
		err   error
	}

	var wg sync.WaitGroup
	primChan := make(chan searchResult, 1)
	secChan := make(chan searchResult, 1)

	// Query primary provider
	wg.Add(1)
	go func() {
		defer wg.Done()
		pCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		works, err := m.primary.Search(pCtx, params)
		primChan <- searchResult{works: works, err: err}
	}()

	// Query secondary provider
	wg.Add(1)
	go func() {
		defer wg.Done()
		sCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		works, err := m.secondary.Search(sCtx, params)
		secChan <- searchResult{works: works, err: err}
	}()

	wg.Wait()
	close(primChan)
	close(secChan)

	pRes := <-primChan
	sRes := <-secChan

	// If both failed, return composite error
	if pRes.err != nil && sRes.err != nil {
		return nil, fmt.Errorf("both providers failed: primary=%v, secondary=%v", pRes.err, sRes.err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	// Index secondary results for metadata enrichment
	secondaryMap := make(map[string]domain.Work)
	for _, w := range sRes.works {
		k := normalizeBookKey(w)
		secondaryMap[k] = w
	}

	var combined []domain.Work
	seen := make(map[string]bool)

	// 1. Process primary results, enriching covers & descriptions if missing
	for _, pw := range pRes.works {
		k := normalizeBookKey(pw)
		if secWork, exists := secondaryMap[k]; exists {
			// Enrich cover if primary has none
			if pw.CoverURL == "" && secWork.CoverURL != "" {
				pw.CoverURL = secWork.CoverURL
			}
			// Enrich description if primary has none
			if pw.Description == "" && secWork.Description != "" {
				pw.Description = secWork.Description
			}
		}
		if !seen[k] {
			seen[k] = true
			combined = append(combined, pw)
		}
		if len(combined) >= limit {
			return combined, nil
		}
	}

	// 2. Append remaining secondary results
	for _, sw := range sRes.works {
		k := normalizeBookKey(sw)
		if !seen[k] {
			seen[k] = true
			combined = append(combined, sw)
		}
		if len(combined) >= limit {
			break
		}
	}

	return combined, nil
}

// GetWork fetches a work, checking Open Library or Google Books depending on ID format,
// and gracefully falling back to the other provider if needed.
func (m *MultiProvider) GetWork(ctx context.Context, providerWorkID string) (*domain.Work, error) {
	trimmed := strings.TrimSpace(providerWorkID)
	if trimmed == "" {
		return nil, domain.ErrWorkNotFound
	}

	isOLID := strings.HasPrefix(trimmed, "OL") || strings.Contains(trimmed, "/works/OL")

	var first, second domain.BookProvider
	if isOLID {
		first = m.primary
		second = m.secondary
	} else {
		first = m.secondary
		second = m.primary
	}

	if first != nil {
		work, err := first.GetWork(ctx, trimmed)
		if err == nil {
			return work, nil
		}
		if !errors.Is(err, domain.ErrWorkNotFound) && second == nil {
			return nil, err
		}
	}

	if second != nil {
		return second.GetWork(ctx, trimmed)
	}

	return nil, domain.ErrWorkNotFound
}

// GetEditionByISBN locates an edition and work by ISBN, querying primary with secondary fallback
// and enriching missing covers/descriptions.
func (m *MultiProvider) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
	var ed *domain.Edition
	var work *domain.Work
	var err error

	if m.primary != nil {
		ed, work, err = m.primary.GetEditionByISBN(ctx, isbn)
	}

	if err != nil || ed == nil {
		if m.secondary != nil {
			return m.secondary.GetEditionByISBN(ctx, isbn)
		}
		return nil, nil, err
	}

	// Enrich missing cover or description from secondary provider
	if m.secondary != nil && (ed.CoverURL == "" || (work != nil && work.CoverURL == "")) {
		if secEd, secWork, secErr := m.secondary.GetEditionByISBN(ctx, isbn); secErr == nil {
			if ed.CoverURL == "" && secEd != nil && secEd.CoverURL != "" {
				ed.CoverURL = secEd.CoverURL
			}
			if work != nil && secWork != nil {
				if work.CoverURL == "" && secWork.CoverURL != "" {
					work.CoverURL = secWork.CoverURL
				}
				if work.Description == "" && secWork.Description != "" {
					work.Description = secWork.Description
				}
			}
		}
	}

	return ed, work, nil
}

// GetEditionsForWork resolves published editions using the primary provider,
// falling back to secondary if primary yields no editions.
func (m *MultiProvider) GetEditionsForWork(ctx context.Context, providerWorkID string, limit int) ([]domain.Edition, error) {
	if m.primary != nil {
		editions, err := m.primary.GetEditionsForWork(ctx, providerWorkID, limit)
		if err == nil && len(editions) > 0 {
			return editions, nil
		}
	}

	if m.secondary != nil {
		return m.secondary.GetEditionsForWork(ctx, providerWorkID, limit)
	}

	return []domain.Edition{}, nil
}
