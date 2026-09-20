package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/adityapandeydev/imprint/backend/internal/infra/security"
)

type mockUserRepo struct {
	usersByID       map[string]*domain.User
	usersByEmail    map[string]*domain.User
	usersByUsername map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		usersByID:       make(map[string]*domain.User),
		usersByEmail:    make(map[string]*domain.User),
		usersByUsername: make(map[string]*domain.User),
	}
}

func (m *mockUserRepo) CreateUser(ctx context.Context, email, username, displayName, passwordHash string) (*domain.User, error) {
	if _, exists := m.usersByEmail[email]; exists {
		return nil, domain.ErrEmailAlreadyExists
	}
	if _, exists := m.usersByUsername[username]; exists {
		return nil, domain.ErrUsernameAlreadyExists
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
	m.usersByID[u.ID] = u
	m.usersByEmail[email] = u
	m.usersByUsername[username] = u
	return u, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := m.usersByID[id]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	if u, ok := m.usersByUsername[username]; ok {
		return u, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) EnsureDefaultUser(ctx context.Context) (*domain.User, error) {
	return &domain.User{ID: "00000000-0000-0000-0000-000000000001", Email: "reader@imprint.app", Username: "reader"}, nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	if u, ok := m.usersByID[userID]; ok {
		u.PasswordHash = passwordHash
		return nil
	}
	return domain.ErrUserNotFound
}

func (m *mockUserRepo) UpdateProfileVisibility(ctx context.Context, userID string, visibility domain.ProfileVisibility) error {
	if u, ok := m.usersByID[userID]; ok {
		u.ProfileVisibility = visibility
		return nil
	}
	return domain.ErrUserNotFound
}

type mockTokenRepo struct {
	mu            sync.Mutex
	refreshTokens map[string]*domain.RefreshToken
	resetTokens   map[string]*domain.PasswordResetToken
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{
		refreshTokens: make(map[string]*domain.RefreshToken),
		resetTokens:   make(map[string]*domain.PasswordResetToken),
	}
}

func (m *mockTokenRepo) SaveRefreshToken(ctx context.Context, t *domain.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshTokens[t.TokenHash] = t
	return nil
}

func (m *mockTokenRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.refreshTokens[tokenHash]; ok {
		return t, nil
	}
	return nil, domain.ErrNotFound
}

func (m *mockTokenRepo) RevokeRefreshToken(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.refreshTokens {
		if t.ID == id {
			t.IsRevoked = true
		}
	}
	return nil
}

func (m *mockTokenRepo) RevokeTokenFamily(ctx context.Context, familyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.refreshTokens {
		if t.FamilyID == familyID {
			t.IsRevoked = true
		}
	}
	return nil
}

func (m *mockTokenRepo) RevokeUserTokens(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.refreshTokens {
		if t.UserID == userID {
			t.IsRevoked = true
		}
	}
	return nil
}

func (m *mockTokenRepo) SavePasswordResetToken(ctx context.Context, t *domain.PasswordResetToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resetTokens[t.TokenHash] = t
	return nil
}

func (m *mockTokenRepo) GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.resetTokens[tokenHash]; ok {
		if !t.IsExpired() {
			return t, nil
		}
	}
	return nil, domain.ErrResetTokenExpired
}

func (m *mockTokenRepo) MarkPasswordResetUsed(ctx context.Context, id string) error {
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

func TestAuthService_RegisterAndLogin(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	jwtSvc := security.NewJWTService("secret-12345", 1*time.Hour)
	authSvc := NewAuthService(userRepo, tokenRepo, jwtSvc)

	// 1. Register
	tokens, err := authSvc.Register(context.Background(), RegisterRequest{
		Email:       "aditya@example.com",
		Username:    "adityap",
		DisplayName: "Aditya Pandey",
		Password:    "supersecret123",
	})
	if err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}
	if tokens.AccessToken == "" {
		t.Errorf("expected non-empty access token")
	}
	if tokens.RefreshToken == "" {
		t.Errorf("expected non-empty refresh token")
	}
	if tokens.User.Email != "aditya@example.com" {
		t.Errorf("expected email aditya@example.com, got %s", tokens.User.Email)
	}

	// 2. Duplicate registration
	_, err = authSvc.Register(context.Background(), RegisterRequest{
		Email:    "aditya@example.com",
		Username: "adityap2",
		Password: "supersecret123",
	})
	if err != domain.ErrEmailAlreadyExists {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}

	// 3. Login with email
	loginTokens, err := authSvc.Login(context.Background(), LoginRequest{
		EmailOrUsername: "aditya@example.com",
		Password:        "supersecret123",
	})
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}
	if loginTokens.AccessToken == "" || loginTokens.RefreshToken == "" {
		t.Errorf("expected valid tokens on login")
	}

	// 4. Login with username
	usernameTokens, err := authSvc.Login(context.Background(), LoginRequest{
		EmailOrUsername: "adityap",
		Password:        "supersecret123",
	})
	if err != nil {
		t.Fatalf("unexpected login with username error: %v", err)
	}
	if usernameTokens.AccessToken == "" {
		t.Errorf("expected valid access token on username login")
	}

	// 5. Login with wrong password
	_, err = authSvc.Login(context.Background(), LoginRequest{
		EmailOrUsername: "adityap",
		Password:        "wrongpass",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// 6. Validate session
	user, err := authSvc.ValidateSession(context.Background(), loginTokens.AccessToken)
	if err != nil {
		t.Fatalf("unexpected session validation error: %v", err)
	}
	if user.ID != tokens.User.ID {
		t.Errorf("expected user ID %s, got %s", tokens.User.ID, user.ID)
	}
}

func TestAuthService_RefreshToken_RotationAndReuse(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	jwtSvc := security.NewJWTService("secret-12345", 15*time.Minute)
	authSvc := NewAuthService(userRepo, tokenRepo, jwtSvc)

	// Register user
	tokens, err := authSvc.Register(context.Background(), RegisterRequest{
		Email:       "rotation@example.com",
		Username:    "rotator",
		DisplayName: "Token Rotator",
		Password:    "supersecret123",
	})
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	oldRefreshToken := tokens.RefreshToken

	// 1. Rotate token
	rotated, err := authSvc.RefreshToken(context.Background(), oldRefreshToken)
	if err != nil {
		t.Fatalf("token rotation failed: %v", err)
	}
	if rotated.RefreshToken == "" || rotated.RefreshToken == oldRefreshToken {
		t.Errorf("expected new distinct refresh token")
	}

	// 2. Token Reuse Detection: Present old already-revoked refresh token
	_, err = authSvc.RefreshToken(context.Background(), oldRefreshToken)
	if err != domain.ErrTokenReused {
		t.Fatalf("expected ErrTokenReused on old token submission, got: %v", err)
	}

	// 3. Verify the rotated token family was completely invalidated
	_, err = authSvc.RefreshToken(context.Background(), rotated.RefreshToken)
	if err != domain.ErrTokenReused && err != domain.ErrInvalidToken {
		t.Fatalf("expected invalidated family token to be rejected, got: %v", err)
	}
}

func TestAuthService_PasswordReset(t *testing.T) {
	userRepo := newMockUserRepo()
	tokenRepo := newMockTokenRepo()
	jwtSvc := security.NewJWTService("secret-12345", 1*time.Hour)
	authSvc := NewAuthService(userRepo, tokenRepo, jwtSvc)

	_, err := authSvc.Register(context.Background(), RegisterRequest{
		Email:       "reset@example.com",
		Username:    "resetuser",
		DisplayName: "Reset User",
		Password:    "oldpassword123",
	})
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Request reset
	resetToken, err := authSvc.RequestPasswordReset(context.Background(), "reset@example.com")
	if err != nil {
		t.Fatalf("password reset request failed: %v", err)
	}
	if resetToken == "" {
		t.Fatal("expected reset token string")
	}

	// Reset password
	err = authSvc.ResetPassword(context.Background(), resetToken, "newpassword123")
	if err != nil {
		t.Fatalf("reset password failed: %v", err)
	}

	// Try logging in with old password -> should fail
	_, err = authSvc.Login(context.Background(), LoginRequest{
		EmailOrUsername: "reset@example.com",
		Password:        "oldpassword123",
	})
	if err != domain.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials with old password, got: %v", err)
	}

	// Login with new password -> should succeed
	newTokens, err := authSvc.Login(context.Background(), LoginRequest{
		EmailOrUsername: "reset@example.com",
		Password:        "newpassword123",
	})
	if err != nil {
		t.Fatalf("login with new password failed: %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("expected valid access token with new password")
	}
}
