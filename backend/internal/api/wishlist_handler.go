package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
	if items == nil {
		items = make([]domain.WishlistItem, 0)
	}

	JSON(w, http.StatusOK, items)
}

// GetTags handles GET /api/v1/wishlist/tags
func (h *WishlistHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == "" {
		Error(w, r, domain.ErrUnauthorized)
		return
	}

	tags, err := h.wishlistService.GetUserTags(r.Context(), userID)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"tags": tags})
}

// AddRequest defines the payload for adding an item to a wishlist.
type AddRequest struct {
	WorkID       string               `json:"work_id"`
	EditionID    *string              `json:"edition_id,omitempty"`
	Status       domain.ReadingStatus `json:"status,omitempty"`
	Priority     int                  `json:"priority,omitempty"`
	Rating       *int                 `json:"rating,omitempty"`
	Tags         []string             `json:"tags,omitempty"`
	Notes        string               `json:"notes,omitempty"`
	Title        string               `json:"title,omitempty"`
	Author       string               `json:"author,omitempty"`
	CoverURL     string               `json:"cover_url,omitempty"`
	OriginalYear *int                 `json:"original_year,omitempty"`
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
		UserID:       userID,
		WorkID:       req.WorkID,
		EditionID:    req.EditionID,
		Status:       req.Status,
		Priority:     req.Priority,
		Rating:       req.Rating,
		Tags:         req.Tags,
		Notes:        req.Notes,
		Title:        req.Title,
		Author:       req.Author,
		CoverURL:     req.CoverURL,
		OriginalYear: req.OriginalYear,
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
	Tags      *[]string             `json:"tags,omitempty"`
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
		Tags:      req.Tags,
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

// ImportGoodreads handles POST /api/v1/wishlist/import/goodreads (capped at 5MB)
func (h *WishlistHandler) ImportGoodreads(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == "" {
		Error(w, r, domain.ErrUnauthorized)
		return
	}

	// 5MB upload limit guardrail
	const maxUploadSize = 5 * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Upload exceeds 5MB limit or is malformed", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing 'file' form field in upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	summary, err := h.wishlistService.ImportGoodreadsCSV(r.Context(), userID, file)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, summary)
}

// Export handles GET /api/v1/wishlist/export?format=csv|json
func (h *WishlistHandler) Export(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == "" {
		Error(w, r, domain.ErrUnauthorized)
		return
	}

	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "" {
		format = "csv"
	}

	data, contentType, err := h.wishlistService.ExportLibrary(r.Context(), userID, format)
	if err != nil {
		Error(w, r, err)
		return
	}

	ext := "csv"
	if format == "json" {
		ext = "json"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"imprint_library.%s\"", ext))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
