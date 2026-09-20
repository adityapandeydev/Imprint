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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/api"
	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/adityapandeydev/imprint/backend/internal/infra/postgres"
	"github.com/adityapandeydev/imprint/backend/internal/infra/provider/composite"
	"github.com/adityapandeydev/imprint/backend/internal/infra/provider/googlebooks"
	"github.com/adityapandeydev/imprint/backend/internal/infra/provider/openlibrary"
	"github.com/adityapandeydev/imprint/backend/internal/infra/security"
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
		db            *postgres.DB
		userRepo      domain.UserRepository
		workRepo      domain.WorkRepository
		editionRepo   domain.EditionRepository
		wishlistRepo  domain.WishlistRepository
		searchRepo    domain.SearchCacheRepository
		tokenRepo     domain.TokenRepository
		analyticsRepo domain.AnalyticsRepository
	)

	dbURL := os.Getenv("DATABASE_URL")
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
			userRepo = postgres.NewUserRepo(db.Pool)
			_, _ = userRepo.EnsureDefaultUser(context.Background())

			workRepo = postgres.NewWorkRepo(db.Pool)
			editionRepo = postgres.NewEditionRepo(db.Pool)
			wishlistRepo = postgres.NewWishlistRepo(db.Pool)
			searchRepo = postgres.NewSearchCacheRepo(db.Pool)
			tokenRepo = postgres.NewPostgresTokenRepo(db.Pool)
			analyticsRepo = postgres.NewAnalyticsRepo(db.Pool)
		}
	} else {
		logger.Info("no live database URL configured in .env - running with in-memory persistence fallback")
	}

	// Fallback to in-memory repos if DB is not connected
	if userRepo == nil {
		userRepo = &memUserRepo{users: make(map[string]*domain.User)}
		_, _ = userRepo.EnsureDefaultUser(context.Background())
	}
	if workRepo == nil {
		workRepo = &memWorkRepo{works: make(map[string]*domain.Work)}
		editionRepo = &memEditionRepo{editions: make(map[string]*domain.Edition)}
		wishlistRepo = &memWishlistRepo{items: make(map[string]*domain.WishlistItem)}
		searchRepo = &memSearchCacheRepo{entries: make(map[string]*domain.SearchCacheEntry)}
		tokenRepo = &memTokenRepo{
			refreshTokens: make(map[string]*domain.RefreshToken),
			resetTokens:   make(map[string]*domain.PasswordResetToken),
		}
		analyticsRepo = &memAnalyticsRepo{items: wishlistRepo, goals: make(map[string]*domain.ReadingGoal)}
	}

	// 4. External Book Provider Setup (Open Library + Google Books Hybrid MultiProvider)
	olClient := openlibrary.NewClient(
		openlibrary.WithBaseURL(os.Getenv("OPEN_LIBRARY_BASE_URL")),
	)
	gbClient := googlebooks.NewClient(
		googlebooks.WithBaseURL(os.Getenv("GOOGLE_BOOKS_BASE_URL")),
		googlebooks.WithAPIKey(os.Getenv("GOOGLE_BOOKS_API_KEY")),
	)
	provider := composite.NewMultiProvider(olClient, gbClient)

	// 5. Application Services & Security Setup
	jwtSvc := security.NewJWTService(os.Getenv("JWT_SECRET"), 15*time.Minute) // 15-minute access tokens
	authSvc := app.NewAuthService(userRepo, tokenRepo, jwtSvc)
	catalogSvc := app.NewCatalogService(provider, workRepo, editionRepo, searchRepo)
	wishlistSvc := app.NewWishlistService(wishlistRepo, catalogSvc, workRepo, editionRepo)
	analyticsSvc := app.NewAnalyticsService(analyticsRepo)
	rateLimiter := security.NewRateLimiter(10 * time.Minute)

	// 6. HTTP Handlers & Router
	authHandler := api.NewAuthHandler(authSvc)
	catalogHandler := api.NewCatalogHandler(catalogSvc)
	wishlistHandler := api.NewWishlistHandler(wishlistSvc)
	analyticsHandler := api.NewAnalyticsHandler(analyticsSvc)
	profileSvc := app.NewProfileService(userRepo, wishlistRepo, analyticsRepo)
	profileHandler := api.NewProfileHandler(profileSvc)

	router := api.NewRouter(api.RouterConfig{
		CatalogHandler:   catalogHandler,
		WishlistHandler:  wishlistHandler,
		AuthHandler:      authHandler,
		AnalyticsHandler: analyticsHandler,
		ProfileHandler:   profileHandler,
		JWTService:       jwtSvc,
		RateLimiter:      rateLimiter,
		Logger:           logger,
		DB:               db,
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
func (m *memWorkRepo) GetWorkByGoogleBooksID(ctx context.Context, gbid string) (*domain.Work, error) {
	for _, w := range m.works {
		if w.GoogleBooksID == gbid {
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
func (m *memEditionRepo) GetEditionByGoogleBooksID(ctx context.Context, gbid string) (*domain.Edition, error) {
	for _, ed := range m.editions {
		if ed.GoogleBooksID == gbid {
			return ed, nil
		}
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
	return m.ListByUserAndTag(ctx, uID, s, nil)
}
func (m *memWishlistRepo) ListByUserAndTag(ctx context.Context, uID string, s *domain.ReadingStatus, tag *string) ([]domain.WishlistItem, error) {
	res := make([]domain.WishlistItem, 0)
	for _, it := range m.items {
		if it.UserID == uID {
			if s != nil && it.Status != *s {
				continue
			}
			if tag != nil && *tag != "" {
				hasTag := false
				for _, t := range it.Tags {
					if strings.EqualFold(t, *tag) {
						hasTag = true
						break
					}
				}
				if !hasTag {
					continue
				}
			}
			res = append(res, *it)
		}
	}
	return res, nil
}
func (m *memWishlistRepo) GetUserTags(ctx context.Context, uID string) ([]domain.TagCount, error) {
	tagMap := make(map[string]int)
	for _, it := range m.items {
		if it.UserID == uID {
			for _, t := range it.Tags {
				tagMap[t]++
			}
		}
	}
	var res []domain.TagCount
	for t, c := range tagMap {
		res = append(res, domain.TagCount{Tag: t, Count: c})
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

type memUserRepo struct {
	users map[string]*domain.User
}

func (m *memUserRepo) CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*domain.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return nil, domain.ErrEmailAlreadyExists
		}
		if strings.EqualFold(u.Username, username) {
			return nil, domain.ErrUsernameAlreadyExists
		}
	}
	u := &domain.User{
		ID:           fmt.Sprintf("usr-%d", time.Now().UnixNano()),
		Email:        email,
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.users[u.ID] = u
	return u, nil
}

func (m *memUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *memUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *memUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Username, username) {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *memUserRepo) EnsureDefaultUser(ctx context.Context) (*domain.User, error) {
	if u, ok := m.users[postgres.DefaultUserID]; ok {
		return u, nil
	}
	u := &domain.User{
		ID:          postgres.DefaultUserID,
		Email:       "reader@imprint.app",
		Username:    "reader",
		DisplayName: "Default Reader",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.users[u.ID] = u
	return u, nil
}

func (m *memUserRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if u, ok := m.users[userID]; ok {
		u.PasswordHash = passwordHash
		u.UpdatedAt = time.Now()
		return nil
	}
	return domain.ErrUserNotFound
}

func (m *memUserRepo) UpdateProfileVisibility(ctx context.Context, userID string, visibility domain.ProfileVisibility) error {
	if u, ok := m.users[userID]; ok {
		u.ProfileVisibility = visibility
		u.UpdatedAt = time.Now()
		return nil
	}
	return domain.ErrUserNotFound
}

type memTokenRepo struct {
	mu            sync.Mutex
	refreshTokens map[string]*domain.RefreshToken
	resetTokens   map[string]*domain.PasswordResetToken
}

func (m *memTokenRepo) SaveRefreshToken(ctx context.Context, t *domain.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshTokens[t.TokenHash] = t
	return nil
}

func (m *memTokenRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.refreshTokens[tokenHash]; ok {
		return t, nil
	}
	return nil, domain.ErrNotFound
}

func (m *memTokenRepo) RevokeRefreshToken(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.refreshTokens {
		if t.ID == id {
			t.IsRevoked = true
		}
	}
	return nil
}

func (m *memTokenRepo) RevokeTokenFamily(ctx context.Context, familyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.refreshTokens {
		if t.FamilyID == familyID {
			t.IsRevoked = true
		}
	}
	return nil
}

func (m *memTokenRepo) RevokeUserTokens(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.refreshTokens {
		if t.UserID == userID {
			t.IsRevoked = true
		}
	}
	return nil
}

func (m *memTokenRepo) SavePasswordResetToken(ctx context.Context, t *domain.PasswordResetToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resetTokens[t.TokenHash] = t
	return nil
}

func (m *memTokenRepo) GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.resetTokens[tokenHash]; ok {
		if !t.IsExpired() {
			return t, nil
		}
	}
	return nil, domain.ErrResetTokenExpired
}

func (m *memTokenRepo) MarkPasswordResetUsed(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.resetTokens {
		if t.ID == id {
			now := time.Now()
			t.UsedAt = &now
		}
	}
	return nil
}

type memAnalyticsRepo struct {
	items domain.WishlistRepository
	goals map[string]*domain.ReadingGoal
}

func (m *memAnalyticsRepo) SetReadingGoal(ctx context.Context, userID string, year int, targetBooks int) (*domain.ReadingGoal, error) {
	if m.goals == nil {
		m.goals = make(map[string]*domain.ReadingGoal)
	}
	key := fmt.Sprintf("%s:%d", userID, year)
	goal := &domain.ReadingGoal{
		ID:          key,
		UserID:      userID,
		Year:        year,
		TargetBooks: targetBooks,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.goals[key] = goal
	return goal, nil
}

func (m *memAnalyticsRepo) GetReadingGoal(ctx context.Context, userID string, year int) (*domain.ReadingGoal, error) {
	if m.goals == nil {
		return nil, domain.ErrNotFound
	}
	key := fmt.Sprintf("%s:%d", userID, year)
	if g, ok := m.goals[key]; ok {
		return g, nil
	}
	return nil, domain.ErrNotFound
}

func (m *memAnalyticsRepo) GetReadingStats(ctx context.Context, userID string, year int) (*domain.ReadingStats, error) {
	all, _ := m.items.ListByUser(ctx, userID, nil)
	stats := &domain.ReadingStats{
		CurrentYear:        year,
		TopGenres:          []domain.GenreCount{},
		TopAuthors:         []domain.AuthorCount{},
		FormatDistribution: make(map[string]int),
	}
	stats.TotalBooks = len(all)
	for _, it := range all {
		if it.Status == domain.StatusCurrentlyReading {
			stats.CurrentlyReading++
		} else if it.Status == domain.StatusWantToRead {
			stats.WantToRead++
		} else if it.Status == domain.StatusFinished {
			stats.BooksFinishedYear++
			stats.TotalPagesRead += 320
		}
	}
	target := 0
	if g, err := m.GetReadingGoal(ctx, userID, year); err == nil && g != nil {
		target = g.TargetBooks
	}
	stats.Challenge = &domain.ReadingChallenge{
		Year:             year,
		TargetBooks:      target,
		BooksFinished:    stats.BooksFinishedYear,
		MonthlyProgress:  make([]domain.MonthlyReadingProgress, 12),
		PacingStatus:     "ON_TRACK",
		PacingMessage:    "You are on track!",
	}
	return stats, nil
}

type memSearchCacheRepo struct {
	mu      sync.RWMutex
	entries map[string]*domain.SearchCacheEntry
}

func (m *memSearchCacheRepo) GetCachedQuery(ctx context.Context, query string) (*domain.SearchCacheEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	trimmed := strings.ToLower(strings.TrimSpace(query))
	if entry, ok := m.entries[trimmed]; ok && entry.ExpiresAt.After(time.Now()) {
		entry.HitCount++
		return entry, nil
	}
	return nil, domain.ErrNotFound
}

func (m *memSearchCacheRepo) SaveCachedQuery(ctx context.Context, entry *domain.SearchCacheEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	trimmed := strings.ToLower(strings.TrimSpace(entry.QueryText))
	if trimmed == "" {
		return nil
	}
	ttl := entry.ExpiresAt
	if ttl.IsZero() || ttl.Before(time.Now()) {
		ttl = time.Now().Add(7 * 24 * time.Hour)
	}
	entry.ExpiresAt = ttl
	entry.UpdatedAt = time.Now()
	m.entries[trimmed] = entry
	return nil
}

func (m *memSearchCacheRepo) IncrementHitCount(ctx context.Context, query string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	trimmed := strings.ToLower(strings.TrimSpace(query))
	if entry, ok := m.entries[trimmed]; ok {
		entry.HitCount++
	}
	return nil
}

func (m *memSearchCacheRepo) PruneExpired(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var count int64
	now := time.Now()
	for k, v := range m.entries {
		if v.ExpiresAt.Before(now) {
			delete(m.entries, k)
			count++
		}
	}
	return count, nil
}
