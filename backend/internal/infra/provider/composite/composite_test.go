package composite

import (
	"context"
	"errors"
	"testing"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

type mockProvider struct {
	name          string
	searchFunc    func(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error)
	getWorkFunc   func(ctx context.Context, id string) (*domain.Work, error)
	getISBNFunc   func(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error)
	getEditionsFn func(ctx context.Context, id string, limit int) ([]domain.Edition, error)
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, params)
	}
	return nil, nil
}
func (m *mockProvider) GetWork(ctx context.Context, id string) (*domain.Work, error) {
	if m.getWorkFunc != nil {
		return m.getWorkFunc(ctx, id)
	}
	return nil, domain.ErrWorkNotFound
}
func (m *mockProvider) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
	if m.getISBNFunc != nil {
		return m.getISBNFunc(ctx, isbn)
	}
	return nil, nil, domain.ErrEditionNotFound
}
func (m *mockProvider) GetEditionsForWork(ctx context.Context, id string, limit int) ([]domain.Edition, error) {
	if m.getEditionsFn != nil {
		return m.getEditionsFn(ctx, id, limit)
	}
	return nil, nil
}

func TestMultiProvider_Search_EnrichmentAndDeduplication(t *testing.T) {
	primary := &mockProvider{
		name: "primary",
		searchFunc: func(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
			return []domain.Work{
				{
					ID:       "OL1W",
					Title:    "Dune",
					Authors:  []domain.Author{{Name: "Frank Herbert"}},
					CoverURL: "", // Missing cover
				},
				{
					ID:       "OL2W",
					Title:    "Dune Messiah",
					Authors:  []domain.Author{{Name: "Frank Herbert"}},
					CoverURL: "https://covers.openlibrary.org/b/id/123-L.jpg",
				},
			}, nil
		},
	}

	secondary := &mockProvider{
		name: "secondary",
		searchFunc: func(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
			return []domain.Work{
				{
					ID:          "GB1",
					Title:       "Dune",
					Authors:     []domain.Author{{Name: "Frank Herbert"}},
					CoverURL:    "https://books.google.com/cover-dune.jpg",
					Description: "Epic sci-fi masterpiece.",
				},
				{
					ID:          "GB3",
					Title:       "Children of Dune",
					Authors:     []domain.Author{{Name: "Frank Herbert"}},
					CoverURL:    "https://books.google.com/cover-children.jpg",
					Description: "Third book in series.",
				},
			}, nil
		},
	}

	mp := NewMultiProvider(primary, secondary)
	results, err := mp.Search(context.Background(), domain.ProviderSearchParams{Query: "Dune", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 deduplicated works, got %d", len(results))
	}

	// First work (Dune) should be enriched with Google Books cover and description
	dune := results[0]
	if dune.Title != "Dune" {
		t.Errorf("expected title Dune, got %s", dune.Title)
	}
	if dune.CoverURL != "https://books.google.com/cover-dune.jpg" {
		t.Errorf("expected enriched cover URL, got %s", dune.CoverURL)
	}
	if dune.Description != "Epic sci-fi masterpiece." {
		t.Errorf("expected enriched description, got %s", dune.Description)
	}

	// Third work should be Children of Dune from secondary provider
	children := results[2]
	if children.Title != "Children of Dune" {
		t.Errorf("expected Children of Dune, got %s", children.Title)
	}
}

func TestMultiProvider_Search_FallbackWhenPrimaryFails(t *testing.T) {
	primary := &mockProvider{
		name: "primary",
		searchFunc: func(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
			return nil, errors.New("open library 504 gateway timeout")
		},
	}

	secondary := &mockProvider{
		name: "secondary",
		searchFunc: func(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
			return []domain.Work{
				{
					ID:      "GB1",
					Title:   "1984",
					Authors: []domain.Author{{Name: "George Orwell"}},
				},
			}, nil
		},
	}

	mp := NewMultiProvider(primary, secondary)
	results, err := mp.Search(context.Background(), domain.ProviderSearchParams{Query: "1984", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error on fallback: %v", err)
	}

	if len(results) != 1 || results[0].Title != "1984" {
		t.Fatalf("expected 1 work '1984' from secondary fallback, got %+v", results)
	}
}

func TestMultiProvider_GetEditionByISBN_Fallback(t *testing.T) {
	primary := &mockProvider{
		name: "primary",
		getISBNFunc: func(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
			return nil, nil, domain.ErrEditionNotFound
		},
	}

	isbn13 := "9780441172719"
	secondary := &mockProvider{
		name: "secondary",
		getISBNFunc: func(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
			return &domain.Edition{
				ID:        "ed-gb",
				Title:     "Dune",
				ISBN13:    &isbn13,
				Publisher: "Ace",
			}, &domain.Work{Title: "Dune"}, nil
		},
	}

	mp := NewMultiProvider(primary, secondary)
	ed, work, err := mp.GetEditionByISBN(context.Background(), isbn13)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ed == nil || ed.Publisher != "Ace" {
		t.Errorf("expected edition from secondary provider, got %+v", ed)
	}
	if work == nil || work.Title != "Dune" {
		t.Errorf("expected work from secondary provider, got %+v", work)
	}
}
