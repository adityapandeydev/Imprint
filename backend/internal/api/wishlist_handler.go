package api

import (
	"encoding/json"
	"net/http"

	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

// WishlistHandler exposes endpoints for managing a user's reading collection.
type WishlistHandler struct {
	wishlistService *app.WishlistService
}

// NewWishlistHandler initializes a WishlistHandler.
func NewWishlistHandler(wishlistService *app.WishlistService) *WishlistHandler {
	return &WishlistHandler{wishlistService: wishlistService}
}

// List handles GET /api/v1/wishlist?status={status}
func (h *WishlistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())

	var statusFilter *domain.ReadingStatus
	if s := r.URL.Query().Get("status"); s != "" {
		parsed, err := domain.ValidateReadingStatus(s)
		if err != nil {
			Error(w, r, err)
			return
		}
		statusFilter = &parsed
	}

	items, err := h.wishlistService.ListUserWishlist(r.Context(), userID, statusFilter)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, items)
}

// AddRequest defines the payload for adding an item to a wishlist.
type AddRequest struct {
	WorkID    string               `json:"work_id"`
	EditionID *string              `json:"edition_id,omitempty"`
	Status    domain.ReadingStatus `json:"status,omitempty"`
	Priority  int                  `json:"priority,omitempty"`
	Notes     string               `json:"notes,omitempty"`
}

// Add handles POST /api/v1/wishlist
func (h *WishlistHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())

	var req AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, r, domain.ErrInvalidStatus)
		return
	}

	item, err := h.wishlistService.AddToWishlist(r.Context(), app.AddToWishlistRequest{
		UserID:    userID,
		WorkID:    req.WorkID,
		EditionID: req.EditionID,
		Status:    req.Status,
		Priority:  req.Priority,
		Notes:     req.Notes,
	})
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusCreated, item)
}

// UpdateRequest defines the payload for updating a wishlist entry.
type UpdateRequest struct {
	EditionID *string               `json:"edition_id,omitempty"`
	Status    *domain.ReadingStatus `json:"status,omitempty"`
	Priority  *int                  `json:"priority,omitempty"`
	Rating    *int                  `json:"rating,omitempty"`
	Notes     *string               `json:"notes,omitempty"`
}

// Update handles PATCH /api/v1/wishlist/{id}
func (h *WishlistHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, r, domain.ErrInvalidStatus)
		return
	}

	item, err := h.wishlistService.UpdateWishlistItem(r.Context(), app.UpdateWishlistRequest{
		ID:        id,
		UserID:    userID,
		EditionID: req.EditionID,
		Status:    req.Status,
		Priority:  req.Priority,
		Rating:    req.Rating,
		Notes:     req.Notes,
	})
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, item)
}

// Delete handles DELETE /api/v1/wishlist/{id}
func (h *WishlistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.wishlistService.RemoveFromWishlist(r.Context(), id, userID); err != nil {
		Error(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
