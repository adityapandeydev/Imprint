package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
	"github.com/adityapandeydev/imprint/backend/internal/infra/security"
)

const (
	AuthCookieName   = "auth_token"
	claimsContextKey = contextKey("user_claims")
)

// ExtractToken retrieves the authentication token from either the Authorization header (Bearer)
// or the secure HttpOnly cookie.
func ExtractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	if cookie, err := r.Cookie(AuthCookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}

// RequireAuth strictly enforces authentication on protected routes.
// If a valid token is not presented, it rejects the request immediately with 401 Unauthorized.
func RequireAuth(jwtSvc *security.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := ExtractToken(r)
			if tokenStr == "" {
				Error(w, r, domain.ErrUnauthorized)
				return
			}

			claims, err := jwtSvc.ValidateToken(tokenStr)
			if err != nil {
				Error(w, r, domain.ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims retrieves verified user claims from request context.
func GetUserClaims(ctx context.Context) *security.UserClaims {
	if v, ok := ctx.Value(claimsContextKey).(*security.UserClaims); ok {
		return v
	}
	return nil
}
