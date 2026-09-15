package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// AnalyticsService coordinates reading analytics and literary statistics.
type AnalyticsService struct {
	analyticsRepo domain.AnalyticsRepository
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(analyticsRepo domain.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
	}
}

// GetReadingStats calculates reading velocity and collection insights for the specified reader.
func (s *AnalyticsService) GetReadingStats(ctx context.Context, userID string, year int) (*domain.ReadingStats, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrUnauthorized
	}
	if year <= 0 {
		year = time.Now().Year()
	}
	if s.analyticsRepo == nil {
		return nil, errors.New("analytics repository not configured")
	}

	return s.analyticsRepo.GetReadingStats(ctx, userID, year)
}
