package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/infra/postgres"
	"github.com/adityapandeydev/imprint/backend/internal/infra/security"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterConfig contains dependencies required to assemble the HTTP router.
type RouterConfig struct {
	CatalogHandler   *CatalogHandler
	WishlistHandler  *WishlistHandler
	AuthHandler      *AuthHandler
	AnalyticsHandler *AnalyticsHandler
	ProfileHandler   *ProfileHandler
	JWTService       *security.JWTService
	RateLimiter      *security.RateLimiter
	Logger           *slog.Logger
	DB               *postgres.DB
}

// SecurityHeadersMiddleware sets defensive HTTP response headers.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// NewRouter constructs and configures the HTTP router with routes and middleware.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global Middleware Stack
	r.Use(RequestIDMiddleware)
	r.Use(CORSMiddleware())
	r.Use(SecurityHeadersMiddleware)
	r.Use(middleware.Compress(5))

	if cfg.Logger != nil {
		r.Use(SlogLoggerMiddleware(cfg.Logger))
		r.Use(RecovererMiddleware(cfg.Logger))
	}

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

	// Social crawler / OpenGraph preview routes
	if cfg.ProfileHandler != nil {
		r.Get("/u/{username}", cfg.ProfileHandler.RenderSocialShareHTML)
		r.Get("/u/{username}/shelf/{shelf}", cfg.ProfileHandler.RenderSocialShareHTML)
	}

	// Helper for rate limiter conditional application
	rateLimit := func(limit int, window time.Duration, bucket string) func(http.Handler) http.Handler {
		if cfg.RateLimiter != nil {
			return cfg.RateLimiter.Middleware(limit, window, bucket)
		}
		return func(next http.Handler) http.Handler { return next }
	}

	// API v1 Namespace
	r.Route("/api/v1", func(v1 chi.Router) {
		// 1. Authentication Routes
		if cfg.AuthHandler != nil {
			v1.Route("/auth", func(auth chi.Router) {
				auth.With(rateLimit(10, 10*time.Minute, "register")).Post("/register", cfg.AuthHandler.Register)
				auth.With(rateLimit(15, time.Minute, "login")).Post("/login", cfg.AuthHandler.Login)
				auth.Post("/refresh", cfg.AuthHandler.Refresh)
				auth.With(rateLimit(5, 15*time.Minute, "forgot-password")).Post("/forgot-password", cfg.AuthHandler.ForgotPassword)
				auth.Post("/reset-password", cfg.AuthHandler.ResetPassword)
				auth.Post("/logout", cfg.AuthHandler.Logout)

				// Strictly protected identity route
				if cfg.JWTService != nil {
					auth.With(RequireAuth(cfg.JWTService)).Get("/me", cfg.AuthHandler.Me)
				}
			})
		}

		// 2. Book Catalog Routes
		v1.Route("/books", func(books chi.Router) {
			books.With(rateLimit(60, time.Minute, "search")).Get("/search", cfg.CatalogHandler.Search)
			books.Get("/{id}", cfg.CatalogHandler.GetBook)
		})

		// 3. Edition Routes
		v1.Route("/editions", func(editions chi.Router) {
			editions.Get("/isbn/{isbn}", cfg.CatalogHandler.GetEditionByISBN)
		})

		// 4. Public Reader Profiles & Social Previews
		if cfg.ProfileHandler != nil {
			v1.Route("/public", func(pub chi.Router) {
				pub.Get("/users/{username}", cfg.ProfileHandler.GetPublicProfile)
				pub.Get("/users/{username}/collection", cfg.ProfileHandler.GetPublicCollection)
				pub.Get("/users/{username}/og.svg", cfg.ProfileHandler.GetOpenGraphSVG)
			})
		}

		// 5. User Reading Analytics & Privacy Settings (Protected)
		if cfg.AnalyticsHandler != nil || cfg.ProfileHandler != nil {
			v1.Route("/users", func(users chi.Router) {
				if cfg.JWTService != nil {
					users.Use(RequireAuth(cfg.JWTService))
				}
				if cfg.AnalyticsHandler != nil {
					users.Get("/stats", cfg.AnalyticsHandler.GetStats)
					users.Put("/goals", cfg.AnalyticsHandler.SetGoal)
				}
				if cfg.ProfileHandler != nil {
					users.Put("/privacy", cfg.ProfileHandler.UpdatePrivacy)
				}
			})
		}

		// 6. Wishlist / Collection Routes (Protected)
		v1.Route("/wishlist", func(wl chi.Router) {
			if cfg.JWTService != nil {
				wl.Use(RequireAuth(cfg.JWTService))
			}
			wl.Get("/", cfg.WishlistHandler.List)
			wl.Post("/", cfg.WishlistHandler.Add)
			wl.Get("/tags", cfg.WishlistHandler.GetTags)
			wl.Post("/import/goodreads", cfg.WishlistHandler.ImportGoodreads)
			wl.Get("/export", cfg.WishlistHandler.Export)
			wl.Patch("/{id}", cfg.WishlistHandler.Update)
			wl.Delete("/{id}", cfg.WishlistHandler.Delete)
		})
	})

	return r
}
