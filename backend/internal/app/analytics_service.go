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

// SetReadingGoal creates or updates the user's annual reading challenge target.
func (s *AnalyticsService) SetReadingGoal(ctx context.Context, userID string, year int, targetBooks int) (*domain.ReadingGoal, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrUnauthorized
	}
	if year <= 0 {
		year = time.Now().Year()
	}
	if targetBooks <= 0 {
		return nil, errors.New("target books must be greater than zero")
	}
	if s.analyticsRepo == nil {
		return nil, errors.New("analytics repository not configured")
	}

	return s.analyticsRepo.SetReadingGoal(ctx, userID, year, targetBooks)
}

// GetReadingGoal retrieves the user's annual reading target.
func (s *AnalyticsService) GetReadingGoal(ctx context.Context, userID string, year int) (*domain.ReadingGoal, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrUnauthorized
	}
	if year <= 0 {
		year = time.Now().Year()
	}
	if s.analyticsRepo == nil {
		return nil, errors.New("analytics repository not configured")
	}

	return s.analyticsRepo.GetReadingGoal(ctx, userID, year)
}

