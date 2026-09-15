package domain

import (
	"context"
	"time"
)

// SearchCacheEntry represents a cached search query and its full normalized results.
type SearchCacheEntry struct {
	ID          string    `json:"id"`
	QueryText   string    `json:"query_text"`
	Results     []Work    `json:"results"`
	ResultCount int       `json:"result_count"`
	HitCount    int       `json:"hit_count"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SearchCacheRepository defines persistence operations for search query caching.
type SearchCacheRepository interface {
	GetCachedQuery(ctx context.Context, query string) (*SearchCacheEntry, error)
	SaveCachedQuery(ctx context.Context, entry *SearchCacheEntry) error
	IncrementHitCount(ctx context.Context, query string) error
	PruneExpired(ctx context.Context) (int64, error)
}
