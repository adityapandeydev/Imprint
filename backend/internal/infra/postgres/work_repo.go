package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkRepo implements domain.WorkRepository using PostgreSQL.
type WorkRepo struct {
	pool *pgxpool.Pool
}

// NewWorkRepo initializes a new WorkRepo.
func NewWorkRepo(pool *pgxpool.Pool) *WorkRepo {
	return &WorkRepo{pool: pool}
}

// SaveWork inserts or updates a Work, its associated Authors, and their WorkAuthor links.
func (r *WorkRepo) SaveWork(ctx context.Context, work *domain.Work) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning save work tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Upsert Work
	var (
		query string
		args  []any
	)

	if work.OpenLibraryWorkID != "" {
		query = `
			INSERT INTO works (title, subtitle, original_year, description, cover_url, open_library_work_id, subject_tags)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (open_library_work_id) DO UPDATE SET
				title = EXCLUDED.title,
				description = COALESCE(NULLIF(EXCLUDED.description, ''), works.description),
				cover_url = COALESCE(NULLIF(EXCLUDED.cover_url, ''), works.cover_url),
				updated_at = NOW()
			RETURNING id, created_at, updated_at;
		`
		args = []any{
			work.Title,
			work.Subtitle,
			work.OriginalYear,
			work.Description,
			work.CoverURL,
			work.OpenLibraryWorkID,
			work.SubjectTags,
		}
	} else if work.ID != "" {
		query = `
			INSERT INTO works (id, title, subtitle, original_year, description, cover_url, subject_tags)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				description = COALESCE(NULLIF(EXCLUDED.description, ''), works.description),
				cover_url = COALESCE(NULLIF(EXCLUDED.cover_url, ''), works.cover_url),
				updated_at = NOW()
			RETURNING id, created_at, updated_at;
		`
		args = []any{
			work.ID,
			work.Title,
			work.Subtitle,
			work.OriginalYear,
			work.Description,
			work.CoverURL,
			work.SubjectTags,
		}
	} else {
		query = `
			INSERT INTO works (title, subtitle, original_year, description, cover_url, subject_tags)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at;
		`
		args = []any{
			work.Title,
			work.Subtitle,
			work.OriginalYear,
			work.Description,
			work.CoverURL,
			work.SubjectTags,
		}
	}

	if err := tx.QueryRow(ctx, query, args...).Scan(&work.ID, &work.CreatedAt, &work.UpdatedAt); err != nil {
		return fmt.Errorf("upserting work: %w", err)
	}

	// 2. Upsert Authors and link them
	for i := range work.Authors {
		author := &work.Authors[i]
		var authorID string

		if author.OpenLibraryID != "" {
			err = tx.QueryRow(ctx, `
				INSERT INTO authors (name, bio, open_library_id)
				VALUES ($1, $2, $3)
				ON CONFLICT (open_library_id) DO UPDATE SET
					name = EXCLUDED.name
				RETURNING id, created_at, updated_at;
			`, author.Name, author.Bio, author.OpenLibraryID).Scan(&author.ID, &author.CreatedAt, &author.UpdatedAt)
		} else {
			// Find existing author by exact name or insert
			err = tx.QueryRow(ctx, `
				SELECT id, created_at, updated_at FROM authors WHERE name = $1 LIMIT 1;
			`, author.Name).Scan(&author.ID, &author.CreatedAt, &author.UpdatedAt)

			if errors.Is(err, pgx.ErrNoRows) {
				err = tx.QueryRow(ctx, `
					INSERT INTO authors (name, bio) VALUES ($1, $2)
					RETURNING id, created_at, updated_at;
				`, author.Name, author.Bio).Scan(&author.ID, &author.CreatedAt, &author.UpdatedAt)
			}
		}

		if err != nil {
			return fmt.Errorf("upserting author %q: %w", author.Name, err)
		}
		authorID = author.ID

		// Link in work_authors
		_, err = tx.Exec(ctx, `
			INSERT INTO work_authors (work_id, author_id, role)
			VALUES ($1, $2, 'AUTHOR')
			ON CONFLICT (work_id, author_id, role) DO NOTHING;
		`, work.ID, authorID)
		if err != nil {
			return fmt.Errorf("linking work_author: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetWorkByID retrieves a Work and its Authors by UUID.
func (r *WorkRepo) GetWorkByID(ctx context.Context, id string) (*domain.Work, error) {
	var work domain.Work
	query := `
		SELECT id, title, subtitle, original_year, description, cover_url, 
		       COALESCE(open_library_work_id, ''), subject_tags, created_at, updated_at
		FROM works
		WHERE id = $1;
	`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&work.ID,
		&work.Title,
		&work.Subtitle,
		&work.OriginalYear,
		&work.Description,
		&work.CoverURL,
		&work.OpenLibraryWorkID,
		&work.SubjectTags,
		&work.CreatedAt,
		&work.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWorkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying work by id: %w", err)
	}

	authors, err := r.getAuthorsForWork(ctx, work.ID)
	if err != nil {
		return nil, err
	}
	work.Authors = authors

	return &work, nil
}

// GetWorkByOpenLibraryID retrieves a Work by its Open Library ID (e.g. "OL27479W").
func (r *WorkRepo) GetWorkByOpenLibraryID(ctx context.Context, olid string) (*domain.Work, error) {
	var work domain.Work
	query := `
		SELECT id, title, subtitle, original_year, description, cover_url, 
		       COALESCE(open_library_work_id, ''), subject_tags, created_at, updated_at
		FROM works
		WHERE open_library_work_id = $1;
	`
	err := r.pool.QueryRow(ctx, query, olid).Scan(
		&work.ID,
		&work.Title,
		&work.Subtitle,
		&work.OriginalYear,
		&work.Description,
		&work.CoverURL,
		&work.OpenLibraryWorkID,
		&work.SubjectTags,
		&work.CreatedAt,
		&work.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrWorkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying work by olid: %w", err)
	}

	authors, err := r.getAuthorsForWork(ctx, work.ID)
	if err != nil {
		return nil, err
	}
	work.Authors = authors

	return &work, nil
}

// SearchLocalWorks queries the local database catalog using full-text search.
func (r *WorkRepo) SearchLocalWorks(ctx context.Context, queryString string, limit int) ([]domain.Work, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, title, subtitle, original_year, description, cover_url, 
		       COALESCE(open_library_work_id, ''), subject_tags, created_at, updated_at
		FROM works
		WHERE to_tsvector('english', title) @@ plainto_tsquery('english', $1)
		   OR title ILIKE '%' || $1 || '%'
		ORDER BY ts_rank(to_tsvector('english', title), plainto_tsquery('english', $1)) DESC, created_at DESC
		LIMIT $2;
	`
	rows, err := r.pool.Query(ctx, query, queryString, limit)
	if err != nil {
		return nil, fmt.Errorf("executing local work search: %w", err)
	}
	defer rows.Close()

	works := make([]domain.Work, 0)
	for rows.Next() {
		var w domain.Work
		if err := rows.Scan(
			&w.ID,
			&w.Title,
			&w.Subtitle,
			&w.OriginalYear,
			&w.Description,
			&w.CoverURL,
			&w.OpenLibraryWorkID,
			&w.SubjectTags,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning local work row: %w", err)
		}
		works = append(works, w)
	}

	for i := range works {
		authors, err := r.getAuthorsForWork(ctx, works[i].ID)
		if err == nil {
			works[i].Authors = authors
		}
	}

	return works, nil
}

func (r *WorkRepo) getAuthorsForWork(ctx context.Context, workID string) ([]domain.Author, error) {
	query := `
		SELECT a.id, a.name, COALESCE(a.bio, ''), COALESCE(a.open_library_id, ''), a.created_at, a.updated_at
		FROM authors a
		JOIN work_authors wa ON a.id = wa.author_id
		WHERE wa.work_id = $1
		ORDER BY a.name ASC;
	`
	rows, err := r.pool.Query(ctx, query, workID)
	if err != nil {
		return nil, fmt.Errorf("querying authors for work: %w", err)
	}
	defer rows.Close()

	var authors []domain.Author
	for rows.Next() {
		var a domain.Author
		if err := rows.Scan(&a.ID, &a.Name, &a.Bio, &a.OpenLibraryID, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning author: %w", err)
		}
		authors = append(authors, a)
	}
	return authors, nil
}
