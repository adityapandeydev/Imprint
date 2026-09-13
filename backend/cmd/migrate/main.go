package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/infra/postgres"
	"github.com/joho/godotenv"
)

func main() {
	for _, envPath := range []string{".env", "../.env", filepath.Join("..", "..", ".env")} {
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			break
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := postgres.NewDB(ctx, postgres.Config{URL: dbURL})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Locate migrations directory
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
		// If running from repo root or elsewhere, check relative paths
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			migrationsDir = filepath.Join("backend", "migrations")
		}
	}

	migrator := postgres.NewMigrator(db.Pool)
	if err := migrator.Up(ctx, migrationsDir); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("All database migrations successfully applied.")
}
