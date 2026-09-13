package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/infra/postgres"
	"github.com/go-chi/chi/v5"
)

// RouterConfig contains dependencies required to assemble the HTTP router.
type RouterConfig struct {
	CatalogHandler  *CatalogHandler
	WishlistHandler *WishlistHandler
	Logger          *slog.Logger
	DB              *postgres.DB
}

// NewRouter constructs and configures the HTTP router with routes and middleware.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global Middleware Stack
	r.Use(RequestIDMiddleware)
	r.Use(CORSMiddleware())
	if cfg.Logger != nil {
		r.Use(SlogLoggerMiddleware(cfg.Logger))
		r.Use(RecovererMiddleware(cfg.Logger))
	}
	r.Use(UserContextMiddleware)

	// Liveness & Readiness Probes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "disabled"
		if cfg.DB != nil && cfg.DB.Pool != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := cfg.DB.Pool.Ping(ctx); err != nil {
				Error(w, r, err)
				return
			}
			dbStatus = "connected"
		}
		JSON(w, http.StatusOK, map[string]any{
			"status":    "ready",
			"database":  dbStatus,
			"timestamp": time.Now().UTC(),
		})
	})

	// API v1 Namespace
	r.Route("/api/v1", func(v1 chi.Router) {
		// Book Catalog Routes
		v1.Route("/books", func(books chi.Router) {
			books.Get("/search", cfg.CatalogHandler.Search)
			books.Get("/{id}", cfg.CatalogHandler.GetBook)
		})

		// Edition Routes
		v1.Route("/editions", func(editions chi.Router) {
			editions.Get("/isbn/{isbn}", cfg.CatalogHandler.GetEditionByISBN)
		})

		// Wishlist / Collection Routes
		v1.Route("/wishlist", func(wl chi.Router) {
			wl.Get("/", cfg.WishlistHandler.List)
			wl.Post("/", cfg.WishlistHandler.Add)
			wl.Patch("/{id}", cfg.WishlistHandler.Update)
			wl.Delete("/{id}", cfg.WishlistHandler.Delete)
		})
	})

	return r
}
