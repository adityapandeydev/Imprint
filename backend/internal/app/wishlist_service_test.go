package app

import (
	"context"
	"errors"
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
		item.ID = "generated-wishlist-uuid"
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
	var list []domain.WishlistItem
	for _, item := range r.items {
		if item.UserID == userID {
			if status != nil && item.Status != *status {
				continue
			}
			list = append(list, *item)
		}
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
