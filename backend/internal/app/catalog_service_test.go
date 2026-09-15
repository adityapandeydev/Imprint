package app

import (
	"context"
	"testing"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// In-memory test mocks
type mockProvider struct {
	searchWorks   []domain.Work
	getWorkResult *domain.Work
	getWorkErr    error
	editions      []domain.Edition
}

func (m *mockProvider) Name() string { return "mock" }
func (m *mockProvider) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	return m.searchWorks, nil
}
func (m *mockProvider) GetWork(ctx context.Context, providerWorkID string) (*domain.Work, error) {
	return m.getWorkResult, m.getWorkErr
}
func (m *mockProvider) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
	return nil, nil, domain.ErrEditionNotFound
}
func (m *mockProvider) GetEditionsForWork(ctx context.Context, providerWorkID string, limit int) ([]domain.Edition, error) {
	return m.editions, nil
}

type mockWorkRepo struct {
	worksByID map[string]*domain.Work
	worksByOL map[string]*domain.Work
}

func newMockWorkRepo() *mockWorkRepo {
	return &mockWorkRepo{
		worksByID: make(map[string]*domain.Work),
		worksByOL: make(map[string]*domain.Work),
	}
}

func (r *mockWorkRepo) SaveWork(ctx context.Context, work *domain.Work) error {
	if work.ID == "" {
		work.ID = "generated-work-uuid"
	}
	r.worksByID[work.ID] = work
	if work.OpenLibraryWorkID != "" {
		r.worksByOL[work.OpenLibraryWorkID] = work
	}
	return nil
}

func (r *mockWorkRepo) GetWorkByID(ctx context.Context, id string) (*domain.Work, error) {
	w, ok := r.worksByID[id]
	if !ok {
		return nil, domain.ErrWorkNotFound
	}
	return w, nil
}

func (r *mockWorkRepo) GetWorkByOpenLibraryID(ctx context.Context, olid string) (*domain.Work, error) {
	w, ok := r.worksByOL[olid]
	if !ok {
		return nil, domain.ErrWorkNotFound
	}
	return w, nil
}

func (r *mockWorkRepo) SearchLocalWorks(ctx context.Context, query string, limit int) ([]domain.Work, error) {
	var result []domain.Work
	for _, w := range r.worksByID {
		result = append(result, *w)
	}
	return result, nil
}

type mockEditionRepo struct {
	editionsByID map[string]*domain.Edition
	byWorkID     map[string][]domain.Edition
}

func newMockEditionRepo() *mockEditionRepo {
	return &mockEditionRepo{
		editionsByID: make(map[string]*domain.Edition),
		byWorkID:     make(map[string][]domain.Edition),
	}
}

func (r *mockEditionRepo) SaveEdition(ctx context.Context, ed *domain.Edition) error {
	if ed.ID == "" {
		ed.ID = "generated-edition-uuid"
	}
	r.editionsByID[ed.ID] = ed
	r.byWorkID[ed.WorkID] = append(r.byWorkID[ed.WorkID], *ed)
	return nil
}

func (r *mockEditionRepo) GetEditionByID(ctx context.Context, id string) (*domain.Edition, error) {
	ed, ok := r.editionsByID[id]
	if !ok {
		return nil, domain.ErrEditionNotFound
	}
	return ed, nil
}

func (r *mockEditionRepo) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, error) {
	for _, ed := range r.editionsByID {
		if (ed.ISBN13 != nil && *ed.ISBN13 == isbn) || (ed.ISBN10 != nil && *ed.ISBN10 == isbn) {
			return ed, nil
		}
	}
	return nil, domain.ErrEditionNotFound
}

func (r *mockEditionRepo) GetEditionsByWorkID(ctx context.Context, workID string) ([]domain.Edition, error) {
	return r.byWorkID[workID], nil
}

func TestCatalogService_Search(t *testing.T) {
	workRepo := newMockWorkRepo()
	editionRepo := newMockEditionRepo()
	provider := &mockProvider{
		searchWorks: []domain.Work{
			{Title: "The Hobbit", OpenLibraryWorkID: "OL27479W"},
			{Title: "The Fellowship of the Ring", OpenLibraryWorkID: "OL27480W"},
		},
	}

	service := NewCatalogService(provider, workRepo, editionRepo)

	works, err := service.Search(context.Background(), "Tolkien", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(works) != 2 {
		t.Fatalf("expected 2 works, got %d", len(works))
	}
}

func TestCatalogService_GetWork_LazyPersistence(t *testing.T) {
	workRepo := newMockWorkRepo()
	editionRepo := newMockEditionRepo()
	provider := &mockProvider{
		getWorkResult: &domain.Work{
			Title:             "Dune",
			OpenLibraryWorkID: "OL100W",
		},
		editions: []domain.Edition{
			{Title: "Dune Paperback", Format: domain.FormatPaperback},
		},
	}

	service := NewCatalogService(provider, workRepo, editionRepo)

	// Fetching by OL ID should fetch from provider, then lazily persist to local repository
	work, editions, err := service.GetWork(context.Background(), "OL100W")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if work.Title != "Dune" {
		t.Errorf("expected title Dune, got %s", work.Title)
	}
	if len(editions) != 1 {
		t.Errorf("expected 1 edition saved, got %d", len(editions))
	}

	// Verify it is now saved in local repository
	saved, err := workRepo.GetWorkByOpenLibraryID(context.Background(), "OL100W")
	if err != nil {
		t.Fatalf("expected work to be in local repo: %v", err)
	}
	if saved.Title != "Dune" {
		t.Errorf("expected Dune in local repo, got %s", saved.Title)
	}
}

func TestCatalogService_GetWork_WhenWorkAlreadyCachedWithoutEditions(t *testing.T) {
	workRepo := newMockWorkRepo()
	editionRepo := newMockEditionRepo()

	// Pre-seed work in local repository (simulating Search caching the work without editions)
	_ = workRepo.SaveWork(context.Background(), &domain.Work{
		ID:                "local-work-uuid-123",
		Title:             "By Way of Deception",
		OpenLibraryWorkID: "OL2287934W",
	})

	provider := &mockProvider{
		editions: []domain.Edition{
			{Title: "By Way of Deception (Hardcover)", Format: domain.FormatHardcover},
			{Title: "By Way of Deception (Paperback)", Format: domain.FormatPaperback},
		},
	}

	service := NewCatalogService(provider, workRepo, editionRepo)

	// Call GetWork by local UUID
	work, editions, err := service.GetWork(context.Background(), "local-work-uuid-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if work == nil || work.Title != "By Way of Deception" {
		t.Errorf("expected work By Way of Deception, got %v", work)
	}

	if len(editions) != 2 {
		t.Fatalf("expected 2 editions to be fetched and returned, got %d", len(editions))
	}

	// Verify editions were saved locally in editionRepo under work.ID
	localEditions, err := editionRepo.GetEditionsByWorkID(context.Background(), "local-work-uuid-123")
	if err != nil || len(localEditions) != 2 {
		t.Errorf("expected 2 editions saved in local editionRepo, got %d (err=%v)", len(localEditions), err)
	}
}
