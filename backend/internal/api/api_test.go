package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/adityapandeydev/imprint/backend/internal/infra/security"
)

// In-memory test doubles for HTTP tests
type testProvider struct{}

func (p *testProvider) Name() string { return "test" }
func (p *testProvider) Search(ctx context.Context, params domain.ProviderSearchParams) ([]domain.Work, error) {
	return []domain.Work{
		{ID: "w-1", Title: "The Hobbit", OpenLibraryWorkID: "OL27479W"},
	}, nil
}
func (p *testProvider) GetWork(ctx context.Context, id string) (*domain.Work, error) {
	if id == "not-found" {
		return nil, domain.ErrWorkNotFound
	}
	return &domain.Work{ID: "w-1", Title: "The Hobbit", OpenLibraryWorkID: "OL27479W"}, nil
}
func (p *testProvider) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, *domain.Work, error) {
	return &domain.Edition{ID: "e-1", Title: "The Hobbit", ISBN13: &isbn}, &domain.Work{ID: "w-1", Title: "The Hobbit"}, nil
}
func (p *testProvider) GetEditionsForWork(ctx context.Context, id string, limit int) ([]domain.Edition, error) {
	isbn := "9780261102217"
	return []domain.Edition{
		{ID: "e-1", WorkID: "w-1", Title: "The Hobbit Paperback", ISBN13: &isbn, Format: domain.FormatPaperback},
	}, nil
}

type testWorkRepo struct {
	works map[string]*domain.Work
}

func (r *testWorkRepo) SaveWork(ctx context.Context, w *domain.Work) error {
	r.works[w.ID] = w
	return nil
}
func (r *testWorkRepo) GetWorkByID(ctx context.Context, id string) (*domain.Work, error) {
	if w, ok := r.works[id]; ok {
		return w, nil
	}
	return nil, domain.ErrWorkNotFound
}
func (r *testWorkRepo) GetWorkByOpenLibraryID(ctx context.Context, olid string) (*domain.Work, error) {
	for _, w := range r.works {
		if w.OpenLibraryWorkID == olid {
			return w, nil
		}
	}
	return nil, domain.ErrWorkNotFound
}
func (r *testWorkRepo) SearchLocalWorks(ctx context.Context, query string, limit int) ([]domain.Work, error) {
	return nil, nil
}

type testEditionRepo struct {
	editions map[string]*domain.Edition
}

func (r *testEditionRepo) SaveEdition(ctx context.Context, e *domain.Edition) error {
	if r.editions == nil {
		r.editions = make(map[string]*domain.Edition)
	}
	r.editions[e.ID] = e
	return nil
}
func (r *testEditionRepo) GetEditionByID(ctx context.Context, id string) (*domain.Edition, error) {
	if e, ok := r.editions[id]; ok {
		return e, nil
	}
	return nil, domain.ErrEditionNotFound
}
func (r *testEditionRepo) GetEditionByISBN(ctx context.Context, isbn string) (*domain.Edition, error) {
	for _, e := range r.editions {
		if e.ISBN13 != nil && *e.ISBN13 == isbn {
			return e, nil
		}
	}
	return nil, domain.ErrEditionNotFound
}
func (r *testEditionRepo) GetEditionsByWorkID(ctx context.Context, workID string) ([]domain.Edition, error) {
	var res []domain.Edition
	for _, e := range r.editions {
		if e.WorkID == workID {
			res = append(res, *e)
		}
	}
	return res, nil
}

type testWishlistRepo struct {
	items map[string]*domain.WishlistItem
}

func (r *testWishlistRepo) Save(ctx context.Context, item *domain.WishlistItem) error {
	if item.ID == "" {
		item.ID = "test-item-uuid"
	}
	r.items[item.ID] = item
	return nil
}
func (r *testWishlistRepo) GetByID(ctx context.Context, id string) (*domain.WishlistItem, error) {
	if it, ok := r.items[id]; ok {
		return it, nil
	}
	return nil, domain.ErrWishlistItemNotFound
}
func (r *testWishlistRepo) GetByUserAndWork(ctx context.Context, uID, wID string) (*domain.WishlistItem, error) {
	for _, it := range r.items {
		if it.UserID == uID && it.WorkID == wID {
			return it, nil
		}
	}
	return nil, domain.ErrWishlistItemNotFound
}
func (r *testWishlistRepo) ListByUser(ctx context.Context, uID string, s *domain.ReadingStatus) ([]domain.WishlistItem, error) {
	var res []domain.WishlistItem
	for _, it := range r.items {
		if it.UserID == uID {
			res = append(res, *it)
		}
	}
	return res, nil
}
func (r *testWishlistRepo) GetUserTags(ctx context.Context, uID string) ([]domain.TagCount, error) {
	tagMap := make(map[string]int)
	for _, it := range r.items {
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
func (r *testWishlistRepo) Update(ctx context.Context, item *domain.WishlistItem) error {
	r.items[item.ID] = item
	return nil
}
func (r *testWishlistRepo) Delete(ctx context.Context, id, uID string) error {
	delete(r.items, id)
	return nil
}

type testUserRepo struct {
	users map[string]*domain.User
}

func (r *testUserRepo) CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*domain.User, error) {
	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			return nil, domain.ErrEmailAlreadyExists
		}
		if strings.EqualFold(u.Username, username) {
			return nil, domain.ErrUsernameAlreadyExists
		}
	}
	u := &domain.User{
		ID:           "u-" + username,
		Email:        email,
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	r.users[u.ID] = u
	return u, nil
}
func (r *testUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}
func (r *testUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range r.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}
func (r *testUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, u := range r.users {
		if strings.EqualFold(u.Username, username) {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}
func (r *testUserRepo) EnsureDefaultUser(ctx context.Context) (*domain.User, error) {
	return &domain.User{ID: "00000000-0000-0000-0000-000000000001", Email: "reader@imprint.app", Username: "reader"}, nil
}
func (r *testUserRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if u, ok := r.users[userID]; ok {
		u.PasswordHash = passwordHash
		return nil
	}
	return domain.ErrUserNotFound
}

type testTokenRepo struct {
	refreshTokens map[string]*domain.RefreshToken
	resetTokens   map[string]*domain.PasswordResetToken
}

func (r *testTokenRepo) SaveRefreshToken(ctx context.Context, t *domain.RefreshToken) error {
	r.refreshTokens[t.TokenHash] = t
	return nil
}
func (r *testTokenRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	if t, ok := r.refreshTokens[hash]; ok {
		return t, nil
	}
	return nil, domain.ErrNotFound
}
func (r *testTokenRepo) RevokeRefreshToken(ctx context.Context, id string) error {
	for _, t := range r.refreshTokens {
		if t.ID == id {
			t.IsRevoked = true
		}
	}
	return nil
}
func (r *testTokenRepo) RevokeTokenFamily(ctx context.Context, familyID string) error {
	for _, t := range r.refreshTokens {
		if t.FamilyID == familyID {
			t.IsRevoked = true
		}
	}
	return nil
}
func (r *testTokenRepo) RevokeUserTokens(ctx context.Context, userID string) error {
	for _, t := range r.refreshTokens {
		if t.UserID == userID {
			t.IsRevoked = true
		}
	}
	return nil
}
func (r *testTokenRepo) SavePasswordResetToken(ctx context.Context, t *domain.PasswordResetToken) error {
	r.resetTokens[t.TokenHash] = t
	return nil
}
func (r *testTokenRepo) GetValidPasswordResetToken(ctx context.Context, hash string) (*domain.PasswordResetToken, error) {
	if t, ok := r.resetTokens[hash]; ok && !t.IsExpired() {
		return t, nil
	}
	return nil, domain.ErrResetTokenExpired
}
func (r *testTokenRepo) MarkPasswordResetUsed(ctx context.Context, id string) error {
	for _, t := range r.resetTokens {
		if t.ID == id {
			now := time.Now()
			t.UsedAt = &now
		}
	}
	return nil
}

type testAnalyticsRepo struct{}

func (r *testAnalyticsRepo) GetReadingStats(ctx context.Context, userID string, year int) (*domain.ReadingStats, error) {
	return &domain.ReadingStats{
		CurrentYear:        year,
		TotalBooks:         5,
		BooksFinishedYear:  3,
		TotalPagesRead:     960,
		CurrentlyReading:   1,
		WantToRead:         1,
		AverageRating:      4.5,
		TopGenres:          []domain.GenreCount{{Genre: "Fantasy", Count: 3}},
		TopAuthors:         []domain.AuthorCount{{Author: "J.R.R. Tolkien", Count: 3}},
		FormatDistribution: map[string]int{"Hardcover": 3, "Paperback": 2},
	}, nil
}

func setupTestRouter() (http.Handler, *security.JWTService) {
	workRepo := &testWorkRepo{works: map[string]*domain.Work{
		"w-1": {ID: "w-1", Title: "The Hobbit", OpenLibraryWorkID: "OL27479W"},
	}}
	editionRepo := &testEditionRepo{}
	wishlistRepo := &testWishlistRepo{items: make(map[string]*domain.WishlistItem)}
	userRepo := &testUserRepo{users: make(map[string]*domain.User)}
	tokenRepo := &testTokenRepo{
		refreshTokens: make(map[string]*domain.RefreshToken),
		resetTokens:   make(map[string]*domain.PasswordResetToken),
	}
	analyticsRepo := &testAnalyticsRepo{}
	provider := &testProvider{}

	jwtSvc := security.NewJWTService("test-secret-12345", 2*time.Hour)
	authSvc := app.NewAuthService(userRepo, tokenRepo, jwtSvc)
	catalogSvc := app.NewCatalogService(provider, workRepo, editionRepo)
	wishlistSvc := app.NewWishlistService(wishlistRepo, catalogSvc, workRepo, editionRepo)
	analyticsSvc := app.NewAnalyticsService(analyticsRepo)
	rateLimiter := security.NewRateLimiter(5 * time.Minute)

	authHandler := NewAuthHandler(authSvc)
	catalogHandler := NewCatalogHandler(catalogSvc)
	wishlistHandler := NewWishlistHandler(wishlistSvc)
	analyticsHandler := NewAnalyticsHandler(analyticsSvc)

	router := NewRouter(RouterConfig{
		CatalogHandler:   catalogHandler,
		WishlistHandler:  wishlistHandler,
		AuthHandler:      authHandler,
		AnalyticsHandler: analyticsHandler,
		JWTService:       jwtSvc,
		RateLimiter:      rateLimiter,
	})

	return router, jwtSvc
}

func TestHealthEndpoint(t *testing.T) {
	router, _ := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	if rec.Header().Get("X-Request-ID") == "" {
		t.Errorf("expected X-Request-ID header to be set")
	}

	var resp ResponseEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode json response: %v", err)
	}
}

func TestSearchBooksEndpoint(t *testing.T) {
	router, _ := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/search?q=Hobbit", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp struct {
		Data []domain.Work `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].Title != "The Hobbit" {
		t.Errorf("unexpected search data: %+v", resp.Data)
	}
}

func TestAuthEndpoints_RegisterLoginAndMe(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Register new user
	regPayload := []byte(`{
		"email": "reader@example.com",
		"username": "reader1",
		"display_name": "Avid Reader",
		"password": "Password123!"
	}`)
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regPayload))
	regReq.Header.Set("Content-Type", "application/json")
	regRec := httptest.NewRecorder()
	router.ServeHTTP(regRec, regReq)

	if regRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on register, got %d: %s", regRec.Code, regRec.Body.String())
	}

	var regResp struct {
		Data domain.AuthTokens `json:"data"`
	}
	if err := json.NewDecoder(regRec.Body).Decode(&regResp); err != nil {
		t.Fatalf("failed to decode register response: %v", err)
	}
	token := regResp.Data.AccessToken
	if token == "" {
		t.Fatalf("expected access token in register response")
	}

	// 2. Access /auth/me with valid Bearer token
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	router.ServeHTTP(meRec, meReq)

	if meRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on /auth/me with valid token, got %d: %s", meRec.Code, meRec.Body.String())
	}

	// 3. Access /auth/me without token -> MUST be 401 Unauthorized
	unauthReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	unauthRec := httptest.NewRecorder()
	router.ServeHTTP(unauthRec, unauthReq)

	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized on /auth/me without token, got %d", unauthRec.Code)
	}

	// 4. Login with correct credentials
	loginPayload := []byte(`{
		"email_or_username": "reader@example.com",
		"password": "Password123!"
	}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginPayload))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on login, got %d", loginRec.Code)
	}

	// 5. Login with wrong password
	badLoginPayload := []byte(`{
		"email_or_username": "reader@example.com",
		"password": "WrongPassword!"
	}`)
	badLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLoginPayload))
	badLoginReq.Header.Set("Content-Type", "application/json")
	badLoginRec := httptest.NewRecorder()
	router.ServeHTTP(badLoginRec, badLoginReq)

	if badLoginRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized on bad login, got %d", badLoginRec.Code)
	}
}

func TestWishlistEndpoints_StrictRouteProtection(t *testing.T) {
	router, jwtSvc := setupTestRouter()

	// Generate a token for test user
	user := &domain.User{
		ID:       "user-auth-uuid-999",
		Email:    "protected@example.com",
		Username: "protecteduser",
	}
	token, _, _ := jwtSvc.GenerateToken(user)

	// =========================================================================
	// 1. VERIFY UNAUTHORIZED REQUESTS ARE STRICTLY REJECTED (401 Unauthorized)
	// =========================================================================

	// Unauthenticated GET /wishlist
	unauthGetReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	unauthGetRec := httptest.NewRecorder()
	router.ServeHTTP(unauthGetRec, unauthGetReq)
	if unauthGetRec.Code != http.StatusUnauthorized {
		t.Fatalf("CRITICAL SECURITY FAILURE: expected 401 Unauthorized on unauthenticated GET /wishlist, got %d", unauthGetRec.Code)
	}

	// Unauthenticated POST /wishlist
	unauthPostReq := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader([]byte(`{"work_id":"w-1"}`)))
	unauthPostReq.Header.Set("Content-Type", "application/json")
	unauthPostRec := httptest.NewRecorder()
	router.ServeHTTP(unauthPostRec, unauthPostReq)
	if unauthPostRec.Code != http.StatusUnauthorized {
		t.Fatalf("CRITICAL SECURITY FAILURE: expected 401 Unauthorized on unauthenticated POST /wishlist, got %d", unauthPostRec.Code)
	}

	// Unauthenticated PATCH /wishlist/{id}
	unauthPatchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/wishlist/item-123", bytes.NewReader([]byte(`{"priority":5}`)))
	unauthPatchReq.Header.Set("Content-Type", "application/json")
	unauthPatchRec := httptest.NewRecorder()
	router.ServeHTTP(unauthPatchRec, unauthPatchReq)
	if unauthPatchRec.Code != http.StatusUnauthorized {
		t.Fatalf("CRITICAL SECURITY FAILURE: expected 401 Unauthorized on unauthenticated PATCH /wishlist, got %d", unauthPatchRec.Code)
	}

	// Unauthenticated DELETE /wishlist/{id}
	unauthDelReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist/item-123", nil)
	unauthDelRec := httptest.NewRecorder()
	router.ServeHTTP(unauthDelRec, unauthDelReq)
	if unauthDelRec.Code != http.StatusUnauthorized {
		t.Fatalf("CRITICAL SECURITY FAILURE: expected 401 Unauthorized on unauthenticated DELETE /wishlist, got %d", unauthDelRec.Code)
	}

	// =========================================================================
	// 2. VERIFY AUTHENTICATED REQUESTS SUCCEED (200 / 201)
	// =========================================================================

	// Authenticated POST /wishlist
	authPostReq := httptest.NewRequest(http.MethodPost, "/api/v1/wishlist", bytes.NewReader([]byte(`{"work_id": "w-1", "priority": 4}`)))
	authPostReq.Header.Set("Content-Type", "application/json")
	authPostReq.Header.Set("Authorization", "Bearer "+token)
	authPostRec := httptest.NewRecorder()
	router.ServeHTTP(authPostRec, authPostReq)

	if authPostRec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on authenticated POST /wishlist, got %d: %s", authPostRec.Code, authPostRec.Body.String())
	}

	// Authenticated GET /wishlist
	authGetReq := httptest.NewRequest(http.MethodGet, "/api/v1/wishlist", nil)
	authGetReq.Header.Set("Authorization", "Bearer "+token)
	authGetRec := httptest.NewRecorder()
	router.ServeHTTP(authGetRec, authGetReq)

	if authGetRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on authenticated GET /wishlist, got %d: %s", authGetRec.Code, authGetRec.Body.String())
	}

	// Authenticated DELETE /wishlist/{id}
	authDelReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wishlist/test-item-uuid", nil)
	authDelReq.Header.Set("Authorization", "Bearer "+token)
	authDelRec := httptest.NewRecorder()
	router.ServeHTTP(authDelRec, authDelReq)

	if authDelRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content on authenticated DELETE, got %d", authDelRec.Code)
	}
}

func TestGetBookEndpoint_WithEditions(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Search first (which caches the work in the repository without editions)
	searchReq := httptest.NewRequest(http.MethodGet, "/api/v1/books/search?q=Hobbit", nil)
	searchRec := httptest.NewRecorder()
	router.ServeHTTP(searchRec, searchReq)
	if searchRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on search, got %d", searchRec.Code)
	}

	// 2. Fetch book details by ID
	bookReq := httptest.NewRequest(http.MethodGet, "/api/v1/books/w-1", nil)
	bookRec := httptest.NewRecorder()
	router.ServeHTTP(bookRec, bookReq)

	if bookRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /books/w-1, got %d: %s", bookRec.Code, bookRec.Body.String())
	}

	var resp struct {
		Data struct {
			Work     domain.Work      `json:"work"`
			Editions []domain.Edition `json:"editions"`
		} `json:"data"`
	}
	if err := json.NewDecoder(bookRec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode get book response: %v", err)
	}

	if resp.Data.Work.Title != "The Hobbit" {
		t.Errorf("expected title 'The Hobbit', got '%s'", resp.Data.Work.Title)
	}
	if len(resp.Data.Editions) == 0 {
		t.Errorf("expected published editions to be resolved and returned, got 0")
	}
}
