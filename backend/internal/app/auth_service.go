package app

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/adityapandeydev/imprint/backend/internal/infra/security"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

const (
	refreshTokenLifetime = 14 * 24 * time.Hour
	resetTokenLifetime    = 1 * time.Hour
)

// AuthService orchestrates user registration, credential verification, and session token generation.
type AuthService struct {
	userRepo  domain.UserRepository
	tokenRepo domain.TokenRepository
	jwtSvc    *security.JWTService
}

// NewAuthService initializes a new AuthService.
func NewAuthService(userRepo domain.UserRepository, tokenRepo domain.TokenRepository, jwtSvc *security.JWTService) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSvc:    jwtSvc,
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
	Login           string `json:"login"`
	Password        string `json:"password"`
}

// PasswordResetRequest contains the email to send reset instructions for.
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// ResetPasswordSubmission contains the recovery token and new password.
type ResetPasswordSubmission struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// createSession issues an access token and an RFC 6819 refresh token belonging to a family.
func (s *AuthService) createSession(ctx context.Context, user *domain.User, familyID string) (*domain.AuthTokens, error) {
	// 1. Issue access token
	tokenStr, expiresAt, err := s.jwtSvc.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %w", err)
	}

	var rawRefresh string
	if s.tokenRepo != nil {
		if familyID == "" {
			familyID = security.NewUUID()
		}

		rawToken, tokenHash, err := security.GenerateSecureToken()
		if err != nil {
			return nil, fmt.Errorf("generating refresh token: %w", err)
		}

		rt := &domain.RefreshToken{
			ID:        security.NewUUID(),
			UserID:    user.ID,
			TokenHash: tokenHash,
			FamilyID:  familyID,
			IsRevoked: false,
			ExpiresAt: time.Now().UTC().Add(refreshTokenLifetime),
			CreatedAt: time.Now().UTC(),
		}

		if err := s.tokenRepo.SaveRefreshToken(ctx, rt); err != nil {
			return nil, fmt.Errorf("saving refresh token: %w", err)
		}

		rawRefresh = rawToken
	}

	return &domain.AuthTokens{
		AccessToken:  tokenStr,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
		User:         *user,
	}, nil
}

// Register validates inputs, hashes passwords with bcrypt, persists to PostgreSQL, and issues session tokens.
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

	// 3. Issue tokens
	return s.createSession(ctx, user, "")
}

// Login verifies reader credentials and returns access and refresh session tokens.
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*domain.AuthTokens, error) {
	cleanIdentifier := strings.TrimSpace(req.EmailOrUsername)
	if cleanIdentifier == "" {
		cleanIdentifier = strings.TrimSpace(req.Login)
	}
	if cleanIdentifier == "" || req.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	var user *domain.User
	var err error

	if strings.Contains(cleanIdentifier, "@") {
		user, err = s.userRepo.GetByEmail(ctx, cleanIdentifier)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, cleanIdentifier)
		if err != nil {
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

	return s.createSession(ctx, user, "")
}

// RefreshToken implements RFC 6819 Refresh Token Rotation & Token Family Reuse Detection.
func (s *AuthService) RefreshToken(ctx context.Context, rawRefreshToken string) (*domain.AuthTokens, error) {
	if s.tokenRepo == nil {
		return nil, domain.ErrInvalidToken
	}

	cleanToken := strings.TrimSpace(rawRefreshToken)
	if cleanToken == "" {
		return nil, domain.ErrInvalidToken
	}

	tokenHash := security.HashToken(cleanToken)
	rt, err := s.tokenRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Reuse Detection: If an already-revoked refresh token is re-submitted,
	// an attacker or adversary is likely replaying compromised credentials.
	// Revoke the entire family immediately!
	if rt.IsRevoked {
		_ = s.tokenRepo.RevokeTokenFamily(ctx, rt.FamilyID)
		return nil, domain.ErrTokenReused
	}

	// Check expiration
	if rt.IsExpired() {
		return nil, domain.ErrInvalidToken
	}

	// Revoke the used token as part of rotation
	if err := s.tokenRepo.RevokeRefreshToken(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("rotating refresh token: %w", err)
	}

	// Fetch user
	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Create new session within the SAME family
	return s.createSession(ctx, user, rt.FamilyID)
}

// RequestPasswordReset creates a time-limited token for account recovery.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	if s.tokenRepo == nil {
		return "", errors.New("token repository not configured")
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	if cleanEmail == "" {
		return "", errors.New("email address is required")
	}

	user, err := s.userRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		// Return empty token without error to prevent email enumeration
		return "", nil
	}

	rawToken, tokenHash, err := security.GenerateSecureToken()
	if err != nil {
		return "", fmt.Errorf("generating password reset token: %w", err)
	}

	prt := &domain.PasswordResetToken{
		ID:        security.NewUUID(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(resetTokenLifetime),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.tokenRepo.SavePasswordResetToken(ctx, prt); err != nil {
		return "", fmt.Errorf("saving password reset token: %w", err)
	}

	return rawToken, nil
}

// ResetPassword validates the reset token, updates password, and revokes all active sessions.
func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if s.tokenRepo == nil {
		return errors.New("token repository not configured")
	}

	cleanToken := strings.TrimSpace(rawToken)
	if cleanToken == "" {
		return domain.ErrResetTokenExpired
	}

	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	tokenHash := security.HashToken(cleanToken)
	prt, err := s.tokenRepo.GetValidPasswordResetToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	// 1. Hash new password
	hash, err := security.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	// 2. Update user's password in database
	if err := s.userRepo.UpdatePassword(ctx, prt.UserID, hash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	// 3. Mark reset token as used
	_ = s.tokenRepo.MarkPasswordResetUsed(ctx, prt.ID)

	// 4. Invalidate all active sessions for security
	_ = s.tokenRepo.RevokeUserTokens(ctx, prt.UserID)

	return nil
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
