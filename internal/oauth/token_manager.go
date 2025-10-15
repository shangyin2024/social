package oauth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"social/internal/config"
	"social/internal/storage"
	"social/pkg/errors"
	"social/pkg/logger"
)

// TokenManager handles token operations including refresh
type TokenManager struct {
	config  *config.Config
	storage storage.Storage
	logger  *logger.Logger
}

// NewTokenManager creates a new token manager
func NewTokenManager(cfg *config.Config, storage storage.Storage, logger *logger.Logger) *TokenManager {
	return &TokenManager{
		config:  cfg,
		storage: storage,
		logger:  logger,
	}
}

// GetValidToken retrieves a valid token, refreshing if necessary
// This method ensures the returned token is valid and not expired
func (tm *TokenManager) GetValidToken(ctx context.Context, userID, provider, serverName string) (*oauth2.Token, error) {
	// Get current token from storage
	token, err := tm.storage.GetToken(ctx, userID, provider, serverName)
	if err != nil {
		tm.logger.Error(ctx, err, "token not found", "provider", provider, "user_id", userID, "server_name", serverName)
		return nil, errors.ErrTokenNotFound
	}

	// Check if token is expired or will expire soon (within 5 minutes)
	if tm.isTokenExpired(token) {
		tm.logger.Info(ctx, "token expired, attempting refresh", "provider", provider, "user_id", userID, "server_name", serverName)

		// Attempt to refresh the token
		newToken, err := tm.refreshToken(ctx, userID, provider, serverName, token)
		if err != nil {
			tm.logger.Error(ctx, err, "token refresh failed", "provider", provider, "user_id", userID)
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}

		tm.logger.Info(ctx, "token refreshed successfully", "provider", provider, "user_id", userID, "server_name", serverName)
		return newToken, nil
	}

	// Token is still valid
	tm.logger.Info(ctx, "token is valid", "provider", provider, "user_id", userID, "server_name", serverName)
	return token, nil
}

// refreshToken refreshes an expired token
func (tm *TokenManager) refreshToken(ctx context.Context, userID, provider, serverName string, currentToken *oauth2.Token) (*oauth2.Token, error) {
	// For Instagram, the refresh token is actually the current access token
	if provider == "instagram" {
		if currentToken.AccessToken == "" {
			tm.logger.Error(ctx, errors.ErrTokenNotFound, "instagram access token not found", "provider", provider, "user_id", userID)
			return nil, fmt.Errorf("instagram access token not available")
		}
		// For Instagram, we use the current access token as the refresh token
		currentToken.RefreshToken = currentToken.AccessToken
	} else {
		// Check if refresh token exists for other platforms
		if currentToken.RefreshToken == "" {
			tm.logger.Error(ctx, errors.ErrTokenNotFound, "refresh token not found", "provider", provider, "user_id", userID)
			return nil, fmt.Errorf("refresh token not available")
		}
	}

	// Get OAuth config
	oauthConfig, err := tm.config.GetServerOAuthConfig(provider, serverName, "")
	if err != nil {
		tm.logger.Error(ctx, err, "failed to get OAuth config", "provider", provider, "server_name", serverName)
		return nil, fmt.Errorf("failed to get OAuth config: %w", err)
	}

	// Create OAuth service
	oauthService := NewOAuthService(oauthConfig)

	// Refresh token
	newToken, err := oauthService.RefreshToken(ctx, currentToken.RefreshToken)
	if err != nil {
		tm.logger.Error(ctx, err, "token refresh failed", "provider", provider, "user_id", userID)
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Save new token to storage
	if err := tm.storage.SaveToken(ctx, userID, provider, serverName, newToken); err != nil {
		tm.logger.Error(ctx, err, "failed to save refreshed token", "provider", provider, "user_id", userID)
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return newToken, nil
}

// isTokenExpired checks if a token is expired or will expire soon
func (tm *TokenManager) isTokenExpired(token *oauth2.Token) bool {
	if token == nil {
		return true
	}

	// If no expiry time is set, consider it expired
	if token.Expiry.IsZero() {
		return true
	}

	// Consider token expired if it expires within 5 minutes
	expiryBuffer := 5 * time.Minute
	return time.Now().Add(expiryBuffer).After(token.Expiry)
}

// CreateAuthenticatedClient creates an HTTP client with automatic token refresh
// This method ensures the client always has a valid token
func (tm *TokenManager) CreateAuthenticatedClient(ctx context.Context, userID, provider, serverName string) (*http.Client, error) {
	// Get a valid token (refreshing if necessary)
	token, err := tm.GetValidToken(ctx, userID, provider, serverName)
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	// Get OAuth config
	oauthConfig, err := tm.config.GetServerOAuthConfig(provider, serverName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth config: %w", err)
	}

	// Create OAuth service
	oauthService := NewOAuthService(oauthConfig)

	// Create client with automatic token refresh
	client := oauthService.CreateClient(ctx, token)

	return client, nil
}

// IsTokenValid checks if a token exists and is valid without refreshing
func (tm *TokenManager) IsTokenValid(ctx context.Context, userID, provider, serverName string) (bool, error) {
	token, err := tm.storage.GetToken(ctx, userID, provider, serverName)
	if err != nil {
		return false, nil // Token not found
	}

	return !tm.isTokenExpired(token), nil
}

// ForceRefreshToken forces a token refresh regardless of expiry status
func (tm *TokenManager) ForceRefreshToken(ctx context.Context, userID, provider, serverName string) (*oauth2.Token, error) {
	// Get current token from storage
	token, err := tm.storage.GetToken(ctx, userID, provider, serverName)
	if err != nil {
		tm.logger.Error(ctx, err, "token not found for force refresh", "provider", provider, "user_id", userID, "server_name", serverName)
		return nil, fmt.Errorf("OAuth token not found")
	}

	tm.logger.Info(ctx, "force refreshing token", "provider", provider, "user_id", userID, "server_name", serverName)

	// Force refresh the token
	newToken, err := tm.refreshToken(ctx, userID, provider, serverName, token)
	if err != nil {
		tm.logger.Error(ctx, err, "force token refresh failed", "provider", provider, "user_id", userID)
		return nil, fmt.Errorf("force token refresh failed: %w", err)
	}

	tm.logger.Info(ctx, "force token refresh successful", "provider", provider, "user_id", userID, "server_name", serverName)
	return newToken, nil
}
