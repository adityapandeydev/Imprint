package domain

import "time"

// AuthTokens contains JWT access token information returned upon registration or login.
type AuthTokens struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	User        User      `json:"user"`
}
