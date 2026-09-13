package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/go-chi/chi/v5"
)

// CatalogHandler exposes endpoints for book searching, work inspection, and edition details.
type CatalogHandler struct {
	catalogService *app.CatalogService
}

// NewCatalogHandler initializes a CatalogHandler.
func NewCatalogHandler(catalogService *app.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

// Search handles GET /api/v1/books/search?q={query}&limit={limit}
func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if parsed, err := strconv.Atoi(lStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 50 {
				limit = 50
			}
		}
	}

	works, err := h.catalogService.Search(r.Context(), query, limit)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, works)
}

// GetBook handles GET /api/v1/books/{id}
func (h *CatalogHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	work, editions, err := h.catalogService.GetWork(r.Context(), id)
	if err != nil {
		Error(w, r, err)
		return
	}

	response := struct {
		Work     any `json:"work"`
		Editions any `json:"editions"`
	}{
		Work:     work,
		Editions: editions,
	}

	JSON(w, http.StatusOK, response)
}

// GetEditionByISBN handles GET /api/v1/editions/isbn/{isbn}
func (h *CatalogHandler) GetEditionByISBN(w http.ResponseWriter, r *http.Request) {
	isbn := chi.URLParam(r, "isbn")
	edition, work, err := h.catalogService.GetEditionByISBN(r.Context(), isbn)
	if err != nil {
		Error(w, r, err)
		return
	}

	response := struct {
		Edition any `json:"edition"`
		Work    any `json:"work,omitempty"`
	}{
		Edition: edition,
		Work:    work,
	}

	JSON(w, http.StatusOK, response)
}
