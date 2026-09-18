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

	// 5. Monthly reading velocity progress (Jan through Dec)
	monthlyProgress := make([]domain.MonthlyReadingProgress, 12)
	for i := 0; i < 12; i++ {
		monthlyProgress[i] = domain.MonthlyReadingProgress{
			Month: i + 1,
			Books: 0,
			Pages: 0,
		}
	}

	monthlyQuery := `
		SELECT 
			EXTRACT(MONTH FROM COALESCE(wi.finished_at, wi.updated_at))::INT AS m,
			COUNT(*),
			COALESCE(SUM(COALESCE(e.page_count, 320)), 0)
		FROM wishlist_items wi
		LEFT JOIN editions e ON wi.edition_id = e.id
		WHERE wi.user_id = $1 
		  AND wi.status = 'FINISHED'
		  AND EXTRACT(YEAR FROM COALESCE(wi.finished_at, wi.updated_at)) = $2
		GROUP BY m
		ORDER BY m;
	`
	if monthRows, mErr := r.pool.Query(ctx, monthlyQuery, userID, year); mErr == nil {
		defer monthRows.Close()
		for monthRows.Next() {
			var m, b, p int
			if scanErr := monthRows.Scan(&m, &b, &p); scanErr == nil && m >= 1 && m <= 12 {
				monthlyProgress[m-1].Books = b
				monthlyProgress[m-1].Pages = p
			}
		}
	}

	// 6. Annual reading challenge target & real-time pacing
	targetBooks := 0
	var goal domain.ReadingGoal
	goalQuery := `
		SELECT id, user_id, year, target_books, created_at, updated_at
		FROM reading_goals
		WHERE user_id = $1 AND year = $2;
	`
	if gErr := r.pool.QueryRow(ctx, goalQuery, userID, year).Scan(
		&goal.ID, &goal.UserID, &goal.Year, &goal.TargetBooks, &goal.CreatedAt, &goal.UpdatedAt,
	); gErr == nil {
		targetBooks = goal.TargetBooks
	}

	isLeap := (year%4 == 0 && year%100 != 0) || (year%400 == 0)
	totalDays := 365
	if isLeap {
		totalDays = 366
	}

	currentYear := time.Now().Year()
	daysElapsed := totalDays
	if year == currentYear {
		daysElapsed = time.Now().YearDay()
	} else if year > currentYear {
		daysElapsed = 0
	}

	percentage := 0.0
	if targetBooks > 0 {
		percentage = math.Round((float64(stats.BooksFinishedYear)/float64(targetBooks))*1000) / 10
	}

	expectedFinished := 0.0
	if targetBooks > 0 && totalDays > 0 {
		expectedFinished = math.Round((float64(targetBooks)*(float64(daysElapsed)/float64(totalDays)))*10) / 10
	}

	pacingDiff := stats.BooksFinishedYear - int(math.Round(expectedFinished))
	pacingStatus := "ON_TRACK"
	pacingMsg := "You are right on track!"

	if targetBooks <= 0 {
		pacingStatus = "NOT_SET"
		pacingMsg = "Set a yearly reading goal to track your pace."
	} else if stats.BooksFinishedYear >= targetBooks {
		pacingStatus = "COMPLETED"
		pacingMsg = fmt.Sprintf("Challenge completed! %d of %d books read 🎉", stats.BooksFinishedYear, targetBooks)
	} else if pacingDiff > 0 {
		pacingStatus = "AHEAD"
		if pacingDiff == 1 {
			pacingMsg = "You are 1 book ahead of schedule"
		} else {
			pacingMsg = fmt.Sprintf("You are %d books ahead of schedule", pacingDiff)
		}
	} else if pacingDiff < 0 {
		behind := -pacingDiff
		pacingStatus = "BEHIND"
		if behind == 1 {
			pacingMsg = "You are 1 book behind schedule"
		} else {
			pacingMsg = fmt.Sprintf("You are %d books behind schedule", behind)
		}
	}

	stats.Challenge = &domain.ReadingChallenge{
		Year:             year,
		TargetBooks:      targetBooks,
		BooksFinished:    stats.BooksFinishedYear,
		Percentage:       percentage,
		DaysElapsed:      daysElapsed,
		TotalDays:        totalDays,
		ExpectedFinished: expectedFinished,
		PacingDiff:       pacingDiff,
		PacingStatus:     pacingStatus,
		PacingMessage:    pacingMsg,
		MonthlyProgress:  monthlyProgress,
	}

	return stats, nil
}

// SetReadingGoal persists or updates a user's reading goal target for a specific year.
func (r *AnalyticsRepo) SetReadingGoal(ctx context.Context, userID string, year int, targetBooks int) (*domain.ReadingGoal, error) {
	if year <= 0 {
		year = time.Now().Year()
	}
	if targetBooks <= 0 {
		return nil, fmt.Errorf("target_books must be greater than zero")
	}

	query := `
		INSERT INTO reading_goals (user_id, year, target_books, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (user_id, year)
		DO UPDATE SET target_books = EXCLUDED.target_books, updated_at = NOW()
		RETURNING id, user_id, year, target_books, created_at, updated_at;
	`

	var g domain.ReadingGoal
	err := r.pool.QueryRow(ctx, query, userID, year, targetBooks).Scan(
		&g.ID, &g.UserID, &g.Year, &g.TargetBooks, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upserting reading goal: %w", err)
	}

	return &g, nil
}

// GetReadingGoal fetches a user's reading goal target for a specific year.
func (r *AnalyticsRepo) GetReadingGoal(ctx context.Context, userID string, year int) (*domain.ReadingGoal, error) {
	if year <= 0 {
		year = time.Now().Year()
	}

	query := `
		SELECT id, user_id, year, target_books, created_at, updated_at
		FROM reading_goals
		WHERE user_id = $1 AND year = $2;
	`

	var g domain.ReadingGoal
	err := r.pool.QueryRow(ctx, query, userID, year).Scan(
		&g.ID, &g.UserID, &g.Year, &g.TargetBooks, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("querying reading goal: %w", err)
	}

	return &g, nil
}
