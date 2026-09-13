package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/api"
	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/adityapandeydev/imprint/backend/internal/infra/postgres"
	"github.com/adityapandeydev/imprint/backend/internal/infra/provider/openlibrary"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Automatically load .env from root or current directory
	for _, envPath := range []string{".env", "../.env", filepath.Join("..", "..", ".env")} {
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			break
		}
	}

	// 2. Structured slog logger setup
	logLevel := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("starting imprint backend server",
		slog.String("port", port),
		slog.String("env", os.Getenv("APP_ENV")),
	)

	// 3. PostgreSQL Database Initialization
	var (
		db           *postgres.DB
		workRepo     domain.WorkRepository
		editionRepo  domain.EditionRepository
		wishlistRepo domain.WishlistRepository
	)

	dbURL := os.Getenv("DATABASE_URL")
	// If the user hasn't configured a real Neon URL yet, provide a graceful mock fallback
	isPlaceholderDB := dbURL == "" || dbURL == "postgres://user:password@ep-project.aws.neon.tech/imprint?sslmode=require"

	if !isPlaceholderDB {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		db, err = postgres.NewDB(ctx, postgres.Config{URL: dbURL})
		if err != nil {
			logger.Warn("could not connect to postgres, running in standalone/search-only mode",
				slog.String("error", err.Error()),
			)
		} else {
			defer db.Close()
			logger.Info("connected to postgres database")

			// Run migrations
			migrator := postgres.NewMigrator(db.Pool)
			migrationsDir := "migrations"
			if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
				migrationsDir = filepath.Join("backend", "migrations")
			}
			if err := migrator.Up(context.Background(), migrationsDir); err != nil {
				logger.Error("failed running database migrations", slog.String("error", err.Error()))
			}

			// Ensure default user exists
			userRepo := postgres.NewUserRepo(db.Pool)
			_, _ = userRepo.EnsureDefaultUser(context.Background())

			workRepo = postgres.NewWorkRepo(db.Pool)
			editionRepo = postgres.NewEditionRepo(db.Pool)
			wishlistRepo = postgres.NewWishlistRepo(db.Pool)
		}
	} else {
		logger.Info("no live database URL configured in .env - running with in-memory persistence fallback")
	}

	// Fallback to in-memory repos if DB is not connected
	if workRepo == nil {
		workRepo = &memWorkRepo{works: make(map[string]*domain.Work)}
		editionRepo = &memEditionRepo{editions: make(map[string]*domain.Edition)}
		wishlistRepo = &memWishlistRepo{items: make(map[string]*domain.WishlistItem)}
	}

	// 4. External Book Provider Setup
	provider := openlibrary.NewClient(
		openlibrary.WithBaseURL(os.Getenv("OPEN_LIBRARY_BASE_URL")),
	)

	// 5. Application Services Setup
	catalogSvc := app.NewCatalogService(provider, workRepo, editionRepo)
	wishlistSvc := app.NewWishlistService(wishlistRepo, catalogSvc, workRepo, editionRepo)

	// 6. HTTP Handlers & Router
	catalogHandler := api.NewCatalogHandler(catalogSvc)
	wishlistHandler := api.NewWishlistHandler(wishlistSvc)

	router := api.NewRouter(api.RouterConfig{
		CatalogHandler:  catalogHandler,
		WishlistHandler: wishlistHandler,
		Logger:          logger,
		DB:              db,
	})

	// 7. Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("http server listening", slog.String("addr", srv.Addr))
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server terminated unexpectedly", slog.String("error", err.Error()))
		}

	case sig := <-shutdown:
		logger.Info("shutdown signal received, initiating graceful shutdown", slog.String("signal", sig.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("failed graceful server shutdown", slog.String("error", err.Error()))
			_ = srv.Close()
		}
	}

	logger.Info("server stopped gracefully")
}

// In-memory fallbacks for standalone/offline server execution
type memWorkRepo struct{ works map[string]*domain.Work }

func (m *memWorkRepo) SaveWork(ctx context.Context, w *domain.Work) error {
	if w.ID == "" {
		w.ID = fmt.Sprintf("work-%d", time.Now().UnixNano())
	}
	m.works[w.ID] = w
	return nil
}
func (m *memWorkRepo) GetWorkByID(ctx context.Context, id string) (*domain.Work, error) {
	if w, ok := m.works[id]; ok {
		return w, nil
	}
	return nil, domain.ErrWorkNotFound
}
func (m *memWorkRepo) GetWorkByOpenLibraryID(ctx context.Context, olid string) (*domain.Work, error) {
	for _, w := range m.works {
		if w.OpenLibraryWorkID == olid {
			return w, nil
		}
	}
	return nil, domain.ErrWorkNotFound
}
func (m *memWorkRepo) SearchLocalWorks(ctx context.Context, query string, limit int) ([]domain.Work, error) {
	return nil, nil
}

type memEditionRepo struct{ editions map[string]*domain.Edition }

func (m *memEditionRepo) SaveEdition(ctx context.Context, ed *domain.Edition) error {
	if ed.ID == "" {
		ed.ID = fmt.Sprintf("ed-%d", time.Now().UnixNano())
	}
	m.editions[ed.ID] = ed
	return nil
}
func (m *memEditionRepo) GetEditionByID(ctx context.Context, id string) (*domain.Edition, error) {
	if ed, ok := m.editions[id]; ok {
		return ed, nil
	}
	return nil, domain.ErrEditionNotFound
}
func (m *memEditionRepo) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, error) {
	return nil, domain.ErrEditionNotFound
}
func (m *memEditionRepo) GetEditionsByWorkID(ctx context.Context, workID string) ([]domain.Edition, error) {
	return nil, nil
}

type memWishlistRepo struct{ items map[string]*domain.WishlistItem }

func (m *memWishlistRepo) Save(ctx context.Context, item *domain.WishlistItem) error {
	if item.ID == "" {
		item.ID = fmt.Sprintf("wl-%d", time.Now().UnixNano())
	}
	m.items[item.ID] = item
	return nil
}
func (m *memWishlistRepo) GetByID(ctx context.Context, id string) (*domain.WishlistItem, error) {
	if it, ok := m.items[id]; ok {
		return it, nil
	}
	return nil, domain.ErrWishlistItemNotFound
}
func (m *memWishlistRepo) GetByUserAndWork(ctx context.Context, uID, wID string) (*domain.WishlistItem, error) {
	for _, it := range m.items {
		if it.UserID == uID && it.WorkID == wID {
			return it, nil
		}
	}
	return nil, domain.ErrWishlistItemNotFound
}
func (m *memWishlistRepo) ListByUser(ctx context.Context, uID string, s *domain.ReadingStatus) ([]domain.WishlistItem, error) {
	res := make([]domain.WishlistItem, 0)
	for _, it := range m.items {
		if it.UserID == uID {
			if s == nil || it.Status == *s {
				res = append(res, *it)
			}
		}
	}
	return res, nil
}
func (m *memWishlistRepo) Update(ctx context.Context, item *domain.WishlistItem) error {
	m.items[item.ID] = item
	return nil
}
func (m *memWishlistRepo) Delete(ctx context.Context, id, uID string) error {
	if it, ok := m.items[id]; ok && it.UserID == uID {
		delete(m.items, id)
	}
	return nil
}
