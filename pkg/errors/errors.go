package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new application error
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Predefined errors
var (
	ErrInvalidRequest     = NewAppError("INVALID_REQUEST", "Invalid request", http.StatusBadRequest)
	ErrUnauthorized       = NewAppError("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	ErrForbidden          = NewAppError("FORBIDDEN", "Forbidden", http.StatusForbidden)
	ErrNotFound           = NewAppError("NOT_FOUND", "Not found", http.StatusNotFound)
	ErrInternalServer     = NewAppError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrServiceUnavailable = NewAppError("SERVICE_UNAVAILABLE", "Service unavailable", http.StatusServiceUnavailable)

	// OAuth specific errors
	ErrInvalidProvider      = NewAppError("INVALID_PROVIDER", "Invalid OAuth provider", http.StatusBadRequest)
	ErrTokenNotFound        = NewAppError("TOKEN_NOT_FOUND", "OAuth token not found", http.StatusUnauthorized)
	ErrTokenExchange        = NewAppError("TOKEN_EXCHANGE_FAILED", "OAuth token exchange failed", http.StatusInternalServerError)
	ErrInvalidState         = NewAppError("INVALID_STATE", "Invalid OAuth state parameter", http.StatusBadRequest)
	ErrPKCEVerifierNotFound = NewAppError("PKCE_VERIFIER_NOT_FOUND", "PKCE verifier not found or expired", http.StatusBadRequest)
	ErrTokenExpired         = NewAppError("TOKEN_EXPIRED", "OAuth token expired", http.StatusUnauthorized)

	// Platform specific errors
	ErrPlatformNotSupported = NewAppError("PLATFORM_NOT_SUPPORTED", "Platform not supported", http.StatusBadRequest)
	ErrContentRequired      = NewAppError("CONTENT_REQUIRED", "Content is required", http.StatusBadRequest)
	ErrMediaIDRequired      = NewAppError("MEDIA_ID_REQUIRED", "Media ID is required", http.StatusBadRequest)
)

// WrapError wraps an error with additional context
func WrapError(err error, message string) *AppError {
	return &AppError{
		Code:    "WRAPPED_ERROR",
		Message: fmt.Sprintf("%s: %v", message, err),
		Status:  http.StatusInternalServerError,
	}
}
