package domain

import (
	"context"
	"time"
)

// GenreCount represents a literary genre/subject and the number of books tagged with it.
type GenreCount struct {
	Genre string `json:"genre"`
	Count int    `json:"count"`
}

// AuthorCount represents an author and the number of books the user has in their collection.
type AuthorCount struct {
	Author string `json:"author"`
	Count  int    `json:"count"`
}

// MonthlyReadingProgress captures books and pages completed in a specific month (1-12).
type MonthlyReadingProgress struct {
	Month int `json:"month"` // 1 = Jan .. 12 = Dec
	Books int `json:"books"`
	Pages int `json:"pages"`
}

// ReadingGoal represents a user's target for a specific year.
type ReadingGoal struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Year        int       `json:"year"`
	TargetBooks int       `json:"target_books"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReadingChallenge encapsulates the goal, current progress, pacing, and monthly velocity.
type ReadingChallenge struct {
	Year             int                      `json:"year"`
	TargetBooks      int                      `json:"target_books"`
	BooksFinished    int                      `json:"books_finished"`
	Percentage       float64                  `json:"percentage"`
	DaysElapsed      int                      `json:"days_elapsed"`
	TotalDays        int                      `json:"total_days"`
	ExpectedFinished float64                  `json:"expected_finished"`
	PacingDiff       int                      `json:"pacing_diff"`
	PacingStatus     string                   `json:"pacing_status"` // "AHEAD", "ON_TRACK", "BEHIND", "COMPLETED", "NOT_STARTED"
	PacingMessage    string                   `json:"pacing_message"`
	MonthlyProgress  []MonthlyReadingProgress `json:"monthly_progress"`
}

// ReadingStats encapsulates a user's reading analytics and velocity metrics.
type ReadingStats struct {
	TotalBooks         int               `json:"total_books"`
	BooksFinishedYear  int               `json:"books_finished_year"`
	TotalPagesRead     int               `json:"total_pages_read"`
	CurrentlyReading   int               `json:"currently_reading"`
	WantToRead         int               `json:"want_to_read"`
	AverageRating      float64           `json:"average_rating"`
	TopGenres          []GenreCount      `json:"top_genres"`
	TopAuthors         []AuthorCount     `json:"top_authors"`
	FormatDistribution map[string]int    `json:"format_distribution"`
	CurrentYear        int               `json:"current_year"`
	Challenge          *ReadingChallenge `json:"challenge,omitempty"`
}

// AnalyticsRepository defines the persistence contract for querying reading statistics.
type AnalyticsRepository interface {
	GetReadingStats(ctx context.Context, userID string, year int) (*ReadingStats, error)
	SetReadingGoal(ctx context.Context, userID string, year int, targetBooks int) (*ReadingGoal, error)
	GetReadingGoal(ctx context.Context, userID string, year int) (*ReadingGoal, error)
}
