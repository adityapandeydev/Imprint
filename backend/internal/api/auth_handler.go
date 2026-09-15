package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/adityapandeydev/imprint/backend/internal/app"
	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authSvc *app.AuthService
}

// NewAuthHandler initializes a new AuthHandler.
func NewAuthHandler(authSvc *app.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// isSecureRequest checks TLS or reverse-proxy headers for certificate-compatible cookie policies.
func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		return true
	}
	return os.Getenv("APP_ENV") == "production"
}

// setAuthCookie writes a secure, HttpOnly session cookie.
func setAuthCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// clearAuthCookie expires and clears the auth cookie on logout.
func clearAuthCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req app.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, r, domain.ErrNotFound)
		return
	}

	tokens, err := h.authSvc.Register(r.Context(), req)
	if err != nil {
		Error(w, r, err)
		return
	}

	setAuthCookie(w, r, tokens.AccessToken, tokens.ExpiresAt)
	JSON(w, http.StatusCreated, tokens)
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req app.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, r, domain.ErrInvalidCredentials)
		return
	}

	tokens, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		Error(w, r, err)
		return
	}

	setAuthCookie(w, r, tokens.AccessToken, tokens.ExpiresAt)
	JSON(w, http.StatusOK, tokens)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w, r)
	JSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

// Me handles GET /api/v1/auth/me (Strictly Protected)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	if userID == "" {
		Error(w, r, domain.ErrUnauthorized)
		return
	}

	user, err := h.authSvc.GetProfile(r.Context(), userID)
	if err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, user)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	tokenStr := req.RefreshToken
	if tokenStr == "" {
		if cookie, err := r.Cookie("imprint_refresh_token"); err == nil {
			tokenStr = cookie.Value
		}
	}

	if tokenStr == "" {
		Error(w, r, domain.ErrInvalidToken)
		return
	}

	tokens, err := h.authSvc.RefreshToken(r.Context(), tokenStr)
	if err != nil {
		Error(w, r, err)
		return
	}

	setAuthCookie(w, r, tokens.AccessToken, tokens.ExpiresAt)
	JSON(w, http.StatusOK, tokens)
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req app.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, r, domain.ErrNotFound)
		return
	}

	rawToken, err := h.authSvc.RequestPasswordReset(r.Context(), req.Email)
	if err != nil {
		Error(w, r, err)
		return
	}

	res := map[string]string{
		"message": "If that email is registered, recovery instructions have been prepared.",
	}
	if rawToken != "" {
		res["reset_token"] = rawToken
	}

	JSON(w, http.StatusOK, res)
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req app.ResetPasswordSubmission
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, r, domain.ErrInvalidToken)
		return
	}

	if err := h.authSvc.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		Error(w, r, err)
		return
	}

	JSON(w, http.StatusOK, map[string]string{
		"message": "Password successfully updated. Please log in with your new password.",
	})
}

