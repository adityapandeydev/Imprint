package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// AnalyticsHandler serves endpoints for personal reading velocity and statistics.
type AnalyticsHandler struct {
	analyticsSvc *app.AnalyticsService
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(analyticsSvc *app.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsSvc: analyticsSvc}
}

// GetStats handles GET /api/v1/users/stats
func (h *AnalyticsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == "" {
		Error(w, r, domain.ErrUnauthorized)
		return
	}

	year := time.Now().Year()
	if yearParam := r.URL.Query().Get("year"); yearParam != "" {
		if y, err := strconv.Atoi(yearParam); err == nil && y > 1900 && y < 2200 {
			year = y
		}
	}

	stats, err := h.analyticsSvc.GetReadingStats(r.Context(), userID, year)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, stats)
}
