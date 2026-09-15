package postgres

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnalyticsRepo implements domain.AnalyticsRepository using PostgreSQL.
type AnalyticsRepo struct {
	pool *pgxpool.Pool
}

// NewAnalyticsRepo initializes a new AnalyticsRepo.
func NewAnalyticsRepo(pool *pgxpool.Pool) *AnalyticsRepo {
	return &AnalyticsRepo{pool: pool}
}

// GetReadingStats aggregates personal reading statistics for the user.
func (r *AnalyticsRepo) GetReadingStats(ctx context.Context, userID string, year int) (*domain.ReadingStats, error) {
	if year <= 0 {
		year = time.Now().Year()
	}

	stats := &domain.ReadingStats{
		CurrentYear:        year,
		TopGenres:          []domain.GenreCount{},
		TopAuthors:         []domain.AuthorCount{},
		FormatDistribution: make(map[string]int),
	}

	// 1. Overview counts & velocity
	overviewQuery := `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE wi.status = 'CURRENTLY_READING'),
			COUNT(*) FILTER (WHERE wi.status = 'WANT_TO_READ'),
			COUNT(*) FILTER (WHERE wi.status = 'FINISHED' AND EXTRACT(YEAR FROM COALESCE(wi.finished_at, wi.updated_at)) = $2),
			COALESCE(AVG(wi.rating) FILTER (WHERE wi.rating IS NOT NULL AND wi.rating > 0), 0),
			COALESCE(SUM(COALESCE(e.page_count, 320)) FILTER (WHERE wi.status = 'FINISHED'), 0)
		FROM wishlist_items wi
		LEFT JOIN editions e ON wi.edition_id = e.id
		WHERE wi.user_id = $1;
	`
	var rawAvg float64
	err := r.pool.QueryRow(ctx, overviewQuery, userID, year).Scan(
		&stats.TotalBooks,
		&stats.CurrentlyReading,
		&stats.WantToRead,
		&stats.BooksFinishedYear,
		&rawAvg,
		&stats.TotalPagesRead,
	)
	if err != nil {
		return nil, fmt.Errorf("aggregating reading overview: %w", err)
	}
	stats.AverageRating = math.Round(rawAvg*10) / 10

	// 2. Top literary genres
	genreQuery := `
		SELECT unnest(w.subject_tags) AS genre, COUNT(*) AS cnt
		FROM wishlist_items wi
		JOIN works w ON wi.work_id = w.id
		WHERE wi.user_id = $1 AND cardinality(w.subject_tags) > 0
		GROUP BY genre
		ORDER BY cnt DESC
		LIMIT 6;
	`
	genreRows, err := r.pool.Query(ctx, genreQuery, userID)
	if err == nil {
		defer genreRows.Close()
		for genreRows.Next() {
			var gc domain.GenreCount
			if scanErr := genreRows.Scan(&gc.Genre, &gc.Count); scanErr == nil {
				stats.TopGenres = append(stats.TopGenres, gc)
			}
		}
	}

	// 3. Top authors
	authorQuery := `
		SELECT a.name, COUNT(DISTINCT wi.work_id) AS cnt
		FROM wishlist_items wi
		JOIN work_authors wa ON wi.work_id = wa.work_id
		JOIN authors a ON wa.author_id = a.id
		WHERE wi.user_id = $1
		GROUP BY a.id, a.name
		ORDER BY cnt DESC
		LIMIT 5;
	`
	authorRows, err := r.pool.Query(ctx, authorQuery, userID)
	if err == nil {
		defer authorRows.Close()
		for authorRows.Next() {
			var ac domain.AuthorCount
			if scanErr := authorRows.Scan(&ac.Author, &ac.Count); scanErr == nil {
				stats.TopAuthors = append(stats.TopAuthors, ac)
			}
		}
	}

	// 4. Format distribution
	formatQuery := `
		SELECT COALESCE(e.format, 'UNKNOWN') AS fmt, COUNT(*) AS cnt
		FROM wishlist_items wi
		LEFT JOIN editions e ON wi.edition_id = e.id
		WHERE wi.user_id = $1
		GROUP BY fmt;
	`
	formatRows, err := r.pool.Query(ctx, formatQuery, userID)
	if err == nil {
		defer formatRows.Close()
		for formatRows.Next() {
			var fmtName string
			var count int
			if scanErr := formatRows.Scan(&fmtName, &count); scanErr == nil {
				stats.FormatDistribution[fmtName] = count
			}
		}
	}

	return stats, nil
}
