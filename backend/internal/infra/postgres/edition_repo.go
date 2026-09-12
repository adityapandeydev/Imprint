package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EditionRepo implements domain.EditionRepository using PostgreSQL.
type EditionRepo struct {
	pool *pgxpool.Pool
}

// NewEditionRepo initializes a new EditionRepo.
func NewEditionRepo(pool *pgxpool.Pool) *EditionRepo {
	return &EditionRepo{pool: pool}
}

// SaveEdition inserts or updates an Edition record.
func (r *EditionRepo) SaveEdition(ctx context.Context, ed *domain.Edition) error {
	var (
		query string
		args  []any
	)

	if ed.OpenLibraryEditionID != "" {
		query = `
			INSERT INTO editions (
				work_id, title, publisher, publication_date, publication_year,
				page_count, language, format, isbn10, isbn13, asin, cover_url,
				description, open_library_edition_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (open_library_edition_id) DO UPDATE SET
				title = EXCLUDED.title,
				publisher = COALESCE(NULLIF(EXCLUDED.publisher, ''), editions.publisher),
				publication_date = COALESCE(NULLIF(EXCLUDED.publication_date, ''), editions.publication_date),
				page_count = COALESCE(EXCLUDED.page_count, editions.page_count),
				cover_url = COALESCE(NULLIF(EXCLUDED.cover_url, ''), editions.cover_url),
				isbn10 = COALESCE(EXCLUDED.isbn10, editions.isbn10),
				isbn13 = COALESCE(EXCLUDED.isbn13, editions.isbn13),
				asin = COALESCE(EXCLUDED.asin, editions.asin),
				updated_at = NOW()
			RETURNING id, created_at, updated_at;
		`
		args = []any{
			ed.WorkID, ed.Title, ed.Publisher, ed.PublicationDate, ed.PublicationYear,
			ed.PageCount, ed.Language, string(ed.Format), ed.ISBN10, ed.ISBN13, ed.ASIN,
			ed.CoverURL, ed.Description, ed.OpenLibraryEditionID,
		}
	} else if ed.ID != "" {
		query = `
			INSERT INTO editions (
				id, work_id, title, publisher, publication_date, publication_year,
				page_count, language, format, isbn10, isbn13, asin, cover_url,
				description, open_library_edition_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				publisher = COALESCE(NULLIF(EXCLUDED.publisher, ''), editions.publisher),
				publication_date = COALESCE(NULLIF(EXCLUDED.publication_date, ''), editions.publication_date),
				page_count = COALESCE(EXCLUDED.page_count, editions.page_count),
				cover_url = COALESCE(NULLIF(EXCLUDED.cover_url, ''), editions.cover_url),
				updated_at = NOW()
			RETURNING id, created_at, updated_at;
		`
		args = []any{
			ed.ID, ed.WorkID, ed.Title, ed.Publisher, ed.PublicationDate, ed.PublicationYear,
			ed.PageCount, ed.Language, string(ed.Format), ed.ISBN10, ed.ISBN13, ed.ASIN,
			ed.CoverURL, ed.Description, ed.OpenLibraryEditionID,
		}
	} else {
		query = `
			INSERT INTO editions (
				work_id, title, publisher, publication_date, publication_year,
				page_count, language, format, isbn10, isbn13, asin, cover_url,
				description
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id, created_at, updated_at;
		`
		args = []any{
			ed.WorkID, ed.Title, ed.Publisher, ed.PublicationDate, ed.PublicationYear,
			ed.PageCount, ed.Language, string(ed.Format), ed.ISBN10, ed.ISBN13, ed.ASIN,
			ed.CoverURL, ed.Description,
		}
	}

	return r.pool.QueryRow(ctx, query, args...).Scan(&ed.ID, &ed.CreatedAt, &ed.UpdatedAt)
}

// GetEditionByID fetches an Edition by UUID.
func (r *EditionRepo) GetEditionByID(ctx context.Context, id string) (*domain.Edition, error) {
	query := `
		SELECT id, work_id, title, COALESCE(publisher, ''), COALESCE(publication_date, ''),
		       publication_year, page_count, COALESCE(language, 'eng'), format,
		       isbn10, isbn13, asin, COALESCE(cover_url, ''), COALESCE(description, ''),
		       COALESCE(open_library_edition_id, ''), created_at, updated_at
		FROM editions
		WHERE id = $1;
	`
	var (
		ed     domain.Edition
		format string
	)
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ed.ID, &ed.WorkID, &ed.Title, &ed.Publisher, &ed.PublicationDate,
		&ed.PublicationYear, &ed.PageCount, &ed.Language, &format,
		&ed.ISBN10, &ed.ISBN13, &ed.ASIN, &ed.CoverURL, &ed.Description,
		&ed.OpenLibraryEditionID, &ed.CreatedAt, &ed.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEditionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying edition by id: %w", err)
	}

	ed.Format = domain.BookFormat(format)
	return &ed, nil
}

// GetEditionByISBN searches for an edition by normalized ISBN-13, ISBN-10, or ASIN.
func (r *EditionRepo) GetEditionByISBN(ctx context.Context, rawIdentifier string) (*domain.Edition, error) {
	cleaned := domain.CleanIdentifier(rawIdentifier)
	query := `
		SELECT id, work_id, title, COALESCE(publisher, ''), COALESCE(publication_date, ''),
		       publication_year, page_count, COALESCE(language, 'eng'), format,
		       isbn10, isbn13, asin, COALESCE(cover_url, ''), COALESCE(description, ''),
		       COALESCE(open_library_edition_id, ''), created_at, updated_at
		FROM editions
		WHERE isbn13 = $1 OR isbn10 = $1 OR asin = $1
		LIMIT 1;
	`
	var (
		ed     domain.Edition
		format string
	)
	err := r.pool.QueryRow(ctx, query, cleaned).Scan(
		&ed.ID, &ed.WorkID, &ed.Title, &ed.Publisher, &ed.PublicationDate,
		&ed.PublicationYear, &ed.PageCount, &ed.Language, &format,
		&ed.ISBN10, &ed.ISBN13, &ed.ASIN, &ed.CoverURL, &ed.Description,
		&ed.OpenLibraryEditionID, &ed.CreatedAt, &ed.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrEditionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying edition by isbn: %w", err)
	}

	ed.Format = domain.BookFormat(format)
	return &ed, nil
}

// GetEditionsByWorkID retrieves all editions linked to a Work.
func (r *EditionRepo) GetEditionsByWorkID(ctx context.Context, workID string) ([]domain.Edition, error) {
	query := `
		SELECT id, work_id, title, COALESCE(publisher, ''), COALESCE(publication_date, ''),
		       publication_year, page_count, COALESCE(language, 'eng'), format,
		       isbn10, isbn13, asin, COALESCE(cover_url, ''), COALESCE(description, ''),
		       COALESCE(open_library_edition_id, ''), created_at, updated_at
		FROM editions
		WHERE work_id = $1
		ORDER BY publication_year DESC NULLS LAST, title ASC;
	`
	rows, err := r.pool.Query(ctx, query, workID)
	if err != nil {
		return nil, fmt.Errorf("querying editions for work: %w", err)
	}
	defer rows.Close()

	var editions []domain.Edition
	for rows.Next() {
		var (
			ed     domain.Edition
			format string
		)
		if err := rows.Scan(
			&ed.ID, &ed.WorkID, &ed.Title, &ed.Publisher, &ed.PublicationDate,
			&ed.PublicationYear, &ed.PageCount, &ed.Language, &format,
			&ed.ISBN10, &ed.ISBN13, &ed.ASIN, &ed.CoverURL, &ed.Description,
			&ed.OpenLibraryEditionID, &ed.CreatedAt, &ed.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning edition row: %w", err)
		}
		ed.Format = domain.BookFormat(format)
		editions = append(editions, ed)
	}
	return editions, nil
}
