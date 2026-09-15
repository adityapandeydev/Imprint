package domain

import "context"

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

// ReadingStats encapsulates a user's reading analytics and velocity metrics.
type ReadingStats struct {
	TotalBooks         int            `json:"total_books"`
	BooksFinishedYear  int            `json:"books_finished_year"`
	TotalPagesRead     int            `json:"total_pages_read"`
	CurrentlyReading   int            `json:"currently_reading"`
	WantToRead         int            `json:"want_to_read"`
	AverageRating      float64        `json:"average_rating"`
	TopGenres          []GenreCount   `json:"top_genres"`
	TopAuthors         []AuthorCount  `json:"top_authors"`
	FormatDistribution map[string]int `json:"format_distribution"`
	CurrentYear        int            `json:"current_year"`
}

// AnalyticsRepository defines the persistence contract for querying reading statistics.
type AnalyticsRepository interface {
	GetReadingStats(ctx context.Context, userID string, year int) (*ReadingStats, error)
}
