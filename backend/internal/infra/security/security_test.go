package security

import (
	"testing"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretPassw0rd!123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Errorf("expected correct password to verify")
	}

	if CheckPasswordHash("WrongPassw0rd", hash) {
		t.Errorf("expected wrong password to fail verification")
	}
}

func TestJWTService(t *testing.T) {
	jwtSvc := NewJWTService("test-secret-key-12345", 1*time.Hour)

	user := &domain.User{
		ID:          "user-uuid-123",
		Email:       "reader@example.com",
		Username:    "aditya",
		DisplayName: "Aditya",
	}

	// 1. Generate token
	token, exp, err := jwtSvc.GenerateToken(user)
	if err != nil {
		t.Fatalf("unexpected token generation error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token")
	}
	if exp.Before(time.Now()) {
		t.Errorf("expected expiration in future")
	}

	// 2. Validate token
	claims, err := jwtSvc.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected token validation error: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected user ID %s, got %s", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, claims.Email)
	}
	if claims.Username != user.Username {
		t.Errorf("expected username %s, got %s", user.Username, claims.Username)
	}

	// 3. Reject invalid token
	_, err = jwtSvc.ValidateToken(token + "tampered")
	if err == nil {
		t.Errorf("expected error for tampered token")
	}

	// 4. Reject expired token
	expiredSvc := NewJWTService("test-secret-key-12345", -1*time.Second)
	expiredToken, _, _ := expiredSvc.GenerateToken(user)
	_, err = jwtSvc.ValidateToken(expiredToken)
	if err == nil {
		t.Errorf("expected error for expired token")
	}
}
