package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrator runs versioned SQL migrations against PostgreSQL.
type Migrator struct {
	pool *pgxpool.Pool
}

// NewMigrator initializes a Migrator.
func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{pool: pool}
}

// Up applies all pending .up.sql migration files in directory.
func (m *Migrator) Up(ctx context.Context, migrationsDir string) error {
	// 1. Ensure schema_migrations table exists
	_, err := m.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	// 2. Find and sort .up.sql migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory %q: %w", migrationsDir, err)
	}

	var upFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	// 3. Apply unapplied migrations inside a transaction
	for _, file := range upFiles {
		version := strings.Split(file, "_")[0]

		var exists bool
		err := m.pool.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1);
		`, version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("checking migration status for %s: %w", version, err)
		}

		if exists {
			continue // Already applied
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return fmt.Errorf("reading migration file %s: %w", file, err)
		}

		tx, err := m.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning migration transaction for %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("applying migration %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1);`, version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("recording migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("committing migration %s: %w", file, err)
		}

		fmt.Printf("Applied migration: %s\n", file)
	}

	return nil
}
