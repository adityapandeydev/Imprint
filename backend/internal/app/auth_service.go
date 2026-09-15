package app

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/adityapandeydev/imprint/backend/internal/infra/security"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// AuthService orchestrates user registration, credential verification, and session token generation.
type AuthService struct {
	userRepo domain.UserRepository
	jwtSvc   *security.JWTService
}

// NewAuthService initializes a new AuthService.
func NewAuthService(userRepo domain.UserRepository, jwtSvc *security.JWTService) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
	}
}

// RegisterRequest contains parameters for creating a new reader account.
type RegisterRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// LoginRequest contains credentials for authenticating an existing reader.
type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username"`
	Password        string `json:"password"`
}

// Register validates inputs, hashes passwords with bcrypt, persists to PostgreSQL, and issues a signed JWT.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*domain.AuthTokens, error) {
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	cleanUsername := strings.ToLower(strings.TrimSpace(req.Username))
	cleanName := strings.TrimSpace(req.DisplayName)

	if cleanEmail == "" {
		return nil, errors.New("email address is required")
	}
	if _, err := mail.ParseAddress(cleanEmail); err != nil {
		return nil, errors.New("invalid email address format")
	}

	if !usernameRegex.MatchString(cleanUsername) {
		return nil, errors.New("username must be between 3 and 30 characters and contain only letters, numbers, and underscores")
	}

	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters long")
	}

	if cleanName == "" {
		cleanName = cleanUsername
	}

	// 1. Hash password
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// 2. Persist user
	user, err := s.userRepo.CreateUser(ctx, cleanEmail, cleanUsername, cleanName, hash)
	if err != nil {
		return nil, err
	}

	// 3. Issue token
	tokenStr, expiresAt, err := s.jwtSvc.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	return &domain.AuthTokens{
		AccessToken: tokenStr,
		ExpiresAt:   expiresAt,
		User:        *user,
	}, nil
}

// Login verifies reader credentials and returns a signed JWT session token.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*domain.AuthTokens, error) {
	cleanIdentifier := strings.TrimSpace(req.EmailOrUsername)
	if cleanIdentifier == "" || req.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	var user *domain.User
	var err error

	// Determine if identifier is an email or username
	if strings.Contains(cleanIdentifier, "@") {
		user, err = s.userRepo.GetByEmail(ctx, cleanIdentifier)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, cleanIdentifier)
		if err != nil {
			// Fallback check by email just in case
			user, err = s.userRepo.GetByEmail(ctx, cleanIdentifier)
		}
	}

	if err != nil || user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Verify bcrypt hash
	if !security.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// Issue token
	tokenStr, expiresAt, err := s.jwtSvc.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generating session token: %w", err)
	}

	return &domain.AuthTokens{
		AccessToken: tokenStr,
		ExpiresAt:   expiresAt,
		User:        *user,
	}, nil
}

// GetProfile retrieves public user account metadata.
func (s *AuthService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, domain.ErrUnauthorized
	}
	return s.userRepo.GetByID(ctx, userID)
}

// ValidateSession validates a token and returns the corresponding user profile.
func (s *AuthService) ValidateSession(ctx context.Context, tokenStr string) (*domain.User, error) {
	claims, err := s.jwtSvc.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetByID(ctx, claims.UserID)
}
