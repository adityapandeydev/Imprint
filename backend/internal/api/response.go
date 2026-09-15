package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/adityapandeydev/imprint/backend/internal/domain"
)

// ResponseEnvelope wraps successful JSON responses.
type ResponseEnvelope struct {
	Data any `json:"data"`
}

// ErrorResponse represents a standardized API error payload.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail holds error diagnostics.
type ErrorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// JSON writes a JSON response with status code and Content-Type header.
func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(ResponseEnvelope{Data: data})
}

// Error maps internal errors to appropriate HTTP status codes and serializes a standard JSON error envelope.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := err.Error()

	switch {
	case errors.Is(err, domain.ErrNotFound),
		errors.Is(err, domain.ErrWorkNotFound),
		errors.Is(err, domain.ErrEditionNotFound),
		errors.Is(err, domain.ErrAuthorNotFound),
		errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrWishlistItemNotFound):
		status = http.StatusNotFound
		code = "NOT_FOUND"

	case errors.Is(err, domain.ErrInvalidISBN),
		errors.Is(err, domain.ErrInvalidASIN),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrInvalidPriority),
		errors.Is(err, domain.ErrInvalidRating),
		errors.Is(err, domain.ErrEmptyTitle):
		status = http.StatusBadRequest
		code = "INVALID_ARGUMENT"

	case errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrInvalidToken),
		errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		code = "UNAUTHORIZED"

	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		code = "FORBIDDEN"

	case errors.Is(err, domain.ErrDuplicateWishlistItem),
		errors.Is(err, domain.ErrEmailAlreadyExists),
		errors.Is(err, domain.ErrUsernameAlreadyExists):
		status = http.StatusConflict
		code = "ALREADY_EXISTS"

	case errors.Is(err, domain.ErrProviderTimeout):
		status = http.StatusGatewayTimeout
		code = "GATEWAY_TIMEOUT"

	case errors.Is(err, domain.ErrProviderUnavailable):
		status = http.StatusBadGateway
		code = "BAD_GATEWAY"
	}

	requestID := GetRequestID(r.Context())

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	})
}
