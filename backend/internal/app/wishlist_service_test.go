package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

type mockWishlistRepo struct {
	items map[string]*domain.WishlistItem
}

func newMockWishlistRepo() *mockWishlistRepo {
	return &mockWishlistRepo{items: make(map[string]*domain.WishlistItem)}
}

func (r *mockWishlistRepo) Save(ctx context.Context, item *domain.WishlistItem) error {
	if item.ID == "" {
		item.ID = fmt.Sprintf("wl-%d", len(r.items)+1)
	}
	r.items[item.ID] = item
	return nil
}

func (r *mockWishlistRepo) GetByID(ctx context.Context, id string) (*domain.WishlistItem, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, domain.ErrWishlistItemNotFound
	}
	return item, nil
}

func (r *mockWishlistRepo) GetByUserAndWork(ctx context.Context, userID, workID string) (*domain.WishlistItem, error) {
	for _, item := range r.items {
		if item.UserID == userID && item.WorkID == workID {
			return item, nil
		}
	}
	return nil, domain.ErrWishlistItemNotFound
}

func (r *mockWishlistRepo) ListByUser(ctx context.Context, userID string, status *domain.ReadingStatus) ([]domain.WishlistItem, error) {
	return r.ListByUserAndTag(ctx, userID, status, nil)
}

func (r *mockWishlistRepo) ListByUserAndTag(ctx context.Context, userID string, status *domain.ReadingStatus, tag *string) ([]domain.WishlistItem, error) {
	var list []domain.WishlistItem
	for _, item := range r.items {
		if item.UserID == userID {
			if status != nil && item.Status != *status {
				continue
			}
			if tag != nil && *tag != "" {
				matched := false
				for _, t := range item.Tags {
					if strings.EqualFold(t, *tag) {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			list = append(list, *item)
		}
	}
	return list, nil
}

func (r *mockWishlistRepo) GetUserTags(ctx context.Context, userID string) ([]domain.TagCount, error) {
	tagMap := make(map[string]int)
	for _, it := range r.items {
		if it.UserID == userID {
			for _, t := range it.Tags {
				tagMap[t]++
			}
		}
	}
	var list []domain.TagCount
	for t, c := range tagMap {
		list = append(list, domain.TagCount{Tag: t, Count: c})
	}
	return list, nil
}

func (r *mockWishlistRepo) Update(ctx context.Context, item *domain.WishlistItem) error {
	r.items[item.ID] = item
	return nil
}

func (r *mockWishlistRepo) Delete(ctx context.Context, id, userID string) error {
	item, ok := r.items[id]
	if !ok || item.UserID != userID {
		return domain.ErrWishlistItemNotFound
	}
	delete(r.items, id)
	return nil
}

func setupTestWishlistService() (*WishlistService, *mockWishlistRepo, *mockWorkRepo) {
	wishlistRepo := newMockWishlistRepo()
	workRepo := newMockWorkRepo()
	editionRepo := newMockEditionRepo()
	provider := &mockProvider{
		getWorkResult: &domain.Work{
			ID:    "work-1",
			Title: "The Hobbit",
		},
	}
	catalogService := NewCatalogService(provider, workRepo, editionRepo)

	// Preload a work
	workRepo.SaveWork(context.Background(), &domain.Work{
		ID:    "work-1",
		Title: "The Hobbit",
	})

	service := NewWishlistService(wishlistRepo, catalogService, workRepo, editionRepo)
	return service, wishlistRepo, workRepo
}

func TestAddToWishlist_Success(t *testing.T) {
	service, _, _ := setupTestWishlistService()

	item, err := service.AddToWishlist(context.Background(), AddToWishlistRequest{
		UserID:   "user-1",
		WorkID:   "work-1",
		Status:   domain.StatusWantToRead,
		Priority: 4,
		Notes:    "Excited to read!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.Status != domain.StatusWantToRead {
		t.Errorf("expected WANT_TO_READ, got %s", item.Status)
	}
	if item.Priority != 4 {
		t.Errorf("expected priority 4, got %d", item.Priority)
	}
	if item.Notes != "Excited to read!" {
		t.Errorf("expected notes to match, got %s", item.Notes)
	}
}

func TestAddToWishlist_DuplicatePrevention(t *testing.T) {
	service, _, _ := setupTestWishlistService()

	_, err := service.AddToWishlist(context.Background(), AddToWishlistRequest{
		UserID: "user-1",
		WorkID: "work-1",
	})
	if err != nil {
		t.Fatalf("initial add failed: %v", err)
	}

	// Attempting to add the same work again should fail
	_, err = service.AddToWishlist(context.Background(), AddToWishlistRequest{
		UserID: "user-1",
		WorkID: "work-1",
	})
	if !errors.Is(err, domain.ErrDuplicateWishlistItem) {
		t.Errorf("expected ErrDuplicateWishlistItem, got %v", err)
	}
}

func TestUpdateWishlistItem_LifecycleTimestamps(t *testing.T) {
	service, _, _ := setupTestWishlistService()

	item, err := service.AddToWishlist(context.Background(), AddToWishlistRequest{
		UserID: "user-1",
		WorkID: "work-1",
		Status: domain.StatusWantToRead,
	})
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}

	if item.StartedAt != nil {
		t.Errorf("expected nil StartedAt initially, got %v", item.StartedAt)
	}

	// Transition to CURRENTLY_READING
	newStatus := domain.StatusCurrentlyReading
	updated, err := service.UpdateWishlistItem(context.Background(), UpdateWishlistRequest{
		ID:     item.ID,
		UserID: "user-1",
		Status: &newStatus,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if updated.StartedAt == nil {
		t.Errorf("expected StartedAt to be populated when transitioning to CURRENTLY_READING")
	}

	// Test invalid priority rejection
	badPriority := 10
	_, err = service.UpdateWishlistItem(context.Background(), UpdateWishlistRequest{
		ID:       item.ID,
		UserID:   "user-1",
		Priority: &badPriority,
	})
	if !errors.Is(err, domain.ErrInvalidPriority) {
		t.Errorf("expected ErrInvalidPriority, got %v", err)
	}
}

func TestRemoveFromWishlist(t *testing.T) {
	service, _, _ := setupTestWishlistService()

	item, _ := service.AddToWishlist(context.Background(), AddToWishlistRequest{
		UserID: "user-1",
		WorkID: "work-1",
	})

	err := service.RemoveFromWishlist(context.Background(), item.ID, "user-1")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Should no longer exist
	list, err := service.ListUserWishlist(context.Background(), "user-1", nil)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list after deletion, got %d items", len(list))
	}
}

func TestWishlistService_Tags(t *testing.T) {
	service, _, _ := setupTestWishlistService()

	item, err := service.AddToWishlist(context.Background(), AddToWishlistRequest{
		UserID: "user-1",
		WorkID: "work-1",
		Tags:   []string{"Sci-Fi", "Favorites"},
	})
	if err != nil {
		t.Fatalf("add with tags failed: %v", err)
	}
	if len(item.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(item.Tags))
	}

	tags, err := service.GetUserTags(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("get user tags failed: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 user tags, got %d", len(tags))
	}
}

func TestWishlistService_GoodreadsImportAndExport(t *testing.T) {
	service, _, _ := setupTestWishlistService()

	csvContent := `Title,Author,ISBN,ISBN13,My Rating,Average Rating,Publisher,Binding,Number of Pages,Year Published,Original Publication Year,Date Read,Date Added,Bookshelves,Exclusive Shelf,My Review
"Dune","Frank Herbert","=""0441172717""","=""9780441172719""","5","4.26","Ace","Paperback","658","1965","1965","2024/05/10","2024/01/01","favorites, sci-fi-classics","read","Masterpiece."
"Neuromancer","William Gibson","=""0441569595""","=""9780441569595""","4","3.9","Ace","Paperback","271","1984","1984","","2024/02/01","cyberpunk","currently-reading",""
"Foundation","Isaac Asimov","=""0553293354""","=""9780553293357""","0","4.05","Spectra","Mass Market Paperback","244","1951","1951","","2024/03/01","to-read","to-read",""
`
	summary, err := service.ImportGoodreadsCSV(context.Background(), "user-goodreads", strings.NewReader(csvContent))
	if err != nil {
		t.Fatalf("goodreads import failed: %v", err)
	}
	if summary.ImportedCount != 3 {
		t.Errorf("expected 3 imported books, got %d (failed=%d, skipped=%d)", summary.ImportedCount, summary.FailedCount, summary.SkippedCount)
	}

	// Verify items and tags
	items, err := service.ListUserWishlist(context.Background(), "user-goodreads", nil)
	if err != nil {
		t.Fatalf("listing imported items failed: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items in wishlist, got %d", len(items))
	}

	// Verify export CSV
	csvData, mime, err := service.ExportLibrary(context.Background(), "user-goodreads", "csv")
	if err != nil {
		t.Fatalf("export csv failed: %v", err)
	}
	if !strings.Contains(mime, "csv") {
		t.Errorf("expected csv mime type, got %s", mime)
	}
	if !strings.Contains(string(csvData), "Dune") || !strings.Contains(string(csvData), "Neuromancer") {
		t.Errorf("expected exported CSV to contain imported titles")
	}

	// Verify export JSON
	jsonData, jsonMime, err := service.ExportLibrary(context.Background(), "user-goodreads", "json")
	if err != nil {
		t.Fatalf("export json failed: %v", err)
	}
	if jsonMime != "application/json" {
		t.Errorf("expected json mime type, got %s", jsonMime)
	}
	if !strings.Contains(string(jsonData), "Foundation") {
		t.Errorf("expected exported JSON to contain Foundation")
	}
}

