package app

import (
	"context"
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

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := security.NewJWTService("secret-12345", 1*time.Hour)
	authSvc := NewAuthService(repo, jwtSvc)

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
	if loginTokens.AccessToken == "" {
		t.Errorf("expected valid access token on login")
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
