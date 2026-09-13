package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// In-memory test doubles for HTTP tests
type testProvider struct{}

func (p *testProvider) Name() string { return "test" }
func (p *testProvider) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	return []domain.Work{
		{ID: "w-1", Title: "The Hobbit", OpenLibraryWorkID: "OL27479W"},
	}, nil
}
func (p *testProvider) GetWork(ctx context.Context, id string) (*domain.Work, error) {
	if id == "not-found" {
		return nil, domain.ErrWorkNotFound
	}
	return &domain.Work{ID: "w-1", Title: "The Hobbit", OpenLibraryWorkID: "OL27479W"}, nil
}
func (p *testProvider) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
	return &domain.Edition{ID: "e-1", Title: "The Hobbit", ISBN13: &isbn}, &domain.Work{ID: "w-1", Title: "The Hobbit"}, nil
}
func (p *testProvider) GetEditionsForWork(ctx context.Context, id string, limit int) ([]domain.Edition, error) {
	isbn := "9780261102217"
	return []domain.Edition{
		{ID: "e-1", WorkID: "w-1", Title: "The Hobbit Paperback", ISBN13: &isbn, Format: domain.FormatPaperback},
	}, nil
}

type testWorkRepo struct {
	works map[string]*domain.Work
}

func (r *testWorkRepo) SaveWork(ctx context.Context, w *domain.Work) error {
	r.works[w.ID] = w
	return nil
}
func (r *testWorkRepo) GetWorkByID(ctx context.Context, id string) (*domain.Work, error) {
	if w, ok := r.works[id]; ok {
		return w, nil
	}
	return nil, domain.ErrWorkNotFound
}
func (r *testWorkRepo) GetWorkByOpenLibraryID(ctx context.Context, olid string) (*domain.Work, error) {
	for _, w := range r.works {
		if w.OpenLibraryWorkID == olid {
			return w, nil
		}
	}
	return nil, domain.ErrWorkNotFound
}
func (r *testWorkRepo) SearchLocalWorks(ctx context.Context, query string, limit int) ([]domain.Work, error) {
	return nil, nil
}

type testEditionRepo struct{}

func (r *testEditionRepo) SaveEdition(ctx context.Context, ed *domain.Edition) error { return nil }
func (r *testEditionRepo) GetEditionByID(ctx context.Context, id string) (*domain.Edition, error) {
	return &domain.Edition{ID: id, Title: "Edition"}, nil
}
func (r *testEditionRepo) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, error) {
	return nil, domain.ErrEditionNotFound
}
func (r *testEditionRepo) GetEditionsByWorkID(ctx context.Context, workID string) ([]domain.Edition, error) {
	return nil, nil
}

type testWishlistRepo struct {
	items map[string]*domain.WishlistItem
}

func (r *testWishlistRepo) Save(ctx context.Context, item *domain.WishlistItem) error {
	if item.ID == "" {
		item.ID = "test-item-uuid"
	}
	r.items[item.ID] = item
	return nil
}
func (r *testWishlistRepo) GetByID(ctx context.Context, id string) (*domain.WishlistItem, error) {
	if it, ok := r.items[id]; ok {
		return it, nil
	}
	return nil, domain.ErrWishlistItemNotFound
}
func (r *testWishlistRepo) GetByUserAndWork(ctx context.Context, uID, wID string) (*domain.WishlistItem, error) {
	return nil, domain.ErrWishlistItemNotFound
}
func (r *testWishlistRepo) ListByUser(ctx context.Context, uID string, s *domain.ReadingStatus) ([]domain.WishlistItem, error) {
	var res []domain.WishlistItem
	for _, it := range r.items {
		res = append(res, *it)
	}
	return res, nil
}
func (r *testWishlistRepo) Update(ctx context.Context, item *domain.WishlistItem) error {
	r.items[item.ID] = item
	return nil
}
func (r *testWishlistRepo) Delete(ctx context.Context, id, uID string) error {
	delete(r.items, id)
	return nil
}

func setupTestRouter() http.Handler {
	workRepo := &testWorkRepo{works: map[string]*domain.Work{
		"w-1": {ID: "w-1", Title: "The Hobbit", OpenLibraryWorkID: "OL27479W"},
	}}
	editionRepo := &testEditionRepo{}
	wishlistRepo := &testWishlistRepo{items: make(map[string]*domain.WishlistItem)}
	provider := &testProvider{}

	catalogSvc := app.NewCatalogService(provider, workRepo, editionRepo)
	wishlistSvc := app.NewWishlistService(wishlistRepo, catalogSvc, workRepo, editionRepo)

	catalogHandler := NewCatalogHandler(catalogSvc)
	wishlistHandler := NewWishlistHandler(wishlistSvc)

	return NewRouter(RouterConfig{
		CatalogHandler:  catalogHandler,
		WishlistHandler: wishlistHandler,
	})
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	if rec.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID header to be set")
	}

	var resp ResponseEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode json response: %v", err)
	}
}

func TestSearchBooksEndpoint(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/search?q=Hobbit", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp struct {
		Data []domain.Work `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].Title != "The Hobbit" {
		t.Errorf("unexpected search data: %+v", resp.Data)
	}
}

func TestGetBookEndpoint_NotFound(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/not-found", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", rec.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error json: %v", err)
	}

	if errResp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", errResp.Error.Code)
	}
}

func TestWishlistLifecycle(t *testing.T) {
	router := setupTestRouter()

	// 1. Add item to wishlist
	payload := []byte(`{"work_id": "w-1", "priority": 4, "notes": "Reading soon"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. List wishlist
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on list, got %d", listRec.Code)
	}

	// 3. Delete from wishlist
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist/test-item-uuid", nil)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", delRec.Code)
	}
}
