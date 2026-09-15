package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SearchCacheRepo implements domain.SearchCacheRepository using PostgreSQL pgxpool.
type SearchCacheRepo struct {
	pool *pgxpool.Pool
}

// NewSearchCacheRepo initializes a new SearchCacheRepo.
func NewSearchCacheRepo(pool *pgxpool.Pool) *SearchCacheRepo {
	return &SearchCacheRepo{pool: pool}
}

// GetCachedQuery retrieves a cached search query and its results if still within TTL.
func (r *SearchCacheRepo) GetCachedQuery(ctx context.Context, query string) (*domain.SearchCacheEntry, error) {
	trimmed := strings.ToLower(strings.TrimSpace(query))
	if trimmed == "" {
		return nil, domain.ErrNotFound
	}

	querySQL := `
		SELECT id, query_text, results, result_count, hit_count, expires_at, created_at, updated_at
		FROM search_queries
		WHERE LOWER(query_text) = $1 AND expires_at > NOW()
	`

	var entry domain.SearchCacheEntry
	var resultsJSON []byte

	err := r.pool.QueryRow(ctx, querySQL, trimmed).Scan(
		&entry.ID,
		&entry.QueryText,
		&resultsJSON,
		&entry.ResultCount,
		&entry.HitCount,
		&entry.ExpiresAt,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("querying search cache: %w", err)
	}

	if err := json.Unmarshal(resultsJSON, &entry.Results); err != nil {
		return nil, fmt.Errorf("decoding cached search results: %w", err)
	}

	// Increment hit count asynchronously without blocking response
	go func() {
		_ = r.IncrementHitCount(context.Background(), trimmed)
	}()

	return &entry, nil
}

// SaveCachedQuery stores or updates a search query with complete results and a 7-day TTL.
func (r *SearchCacheRepo) SaveCachedQuery(ctx context.Context, entry *domain.SearchCacheEntry) error {
	trimmed := strings.ToLower(strings.TrimSpace(entry.QueryText))
	if trimmed == "" {
		return nil
	}

	resultsJSON, err := json.Marshal(entry.Results)
	if err != nil {
		return fmt.Errorf("marshaling results for cache: %w", err)
	}

	ttl := entry.ExpiresAt
	if ttl.IsZero() || ttl.Before(time.Now()) {
		ttl = time.Now().Add(7 * 24 * time.Hour) // 7-day default TTL
	}

	upsertSQL := `
		INSERT INTO search_queries (query_text, results, result_count, hit_count, expires_at, updated_at)
		VALUES ($1, $2, $3, 1, $4, NOW())
		ON CONFLICT (query_text) DO UPDATE SET
			results = EXCLUDED.results,
			result_count = EXCLUDED.result_count,
			hit_count = search_queries.hit_count + 1,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
	`

	_, err = r.pool.Exec(ctx, upsertSQL, trimmed, resultsJSON, len(entry.Results), ttl)
	if err != nil {
		return fmt.Errorf("persisting search cache entry: %w", err)
	}

	return nil
}

// IncrementHitCount bumps the search frequency counter for trending analytics.
func (r *SearchCacheRepo) IncrementHitCount(ctx context.Context, query string) error {
	trimmed := strings.ToLower(strings.TrimSpace(query))
	updateSQL := `UPDATE search_queries SET hit_count = hit_count + 1, updated_at = NOW() WHERE query_text = $1`
	_, err := r.pool.Exec(ctx, updateSQL, trimmed)
	return err
}

// PruneExpired removes all search queries whose TTL has lapsed.
func (r *SearchCacheRepo) PruneExpired(ctx context.Context) (int64, error) {
	deleteSQL := `DELETE FROM search_queries WHERE expires_at < NOW()`
	tag, err := r.pool.Exec(ctx, deleteSQL)
	if err != nil {
		return 0, fmt.Errorf("pruning expired search queries: %w", err)
	}
	return tag.RowsAffected(), nil
}
