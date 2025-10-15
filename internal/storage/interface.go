package storage

import (
	"context"

	"golang.org/x/oauth2"
)

// Storage defines the interface for token and PKCE storage
type Storage interface {
	// Token operations
	SaveToken(ctx context.Context, userID, provider, serverName string, token *oauth2.Token) error
	GetToken(ctx context.Context, userID, provider, serverName string) (*oauth2.Token, error)
	DeleteToken(ctx context.Context, userID, provider, serverName string) error

	// PKCE operations
	SavePKCEVerifier(ctx context.Context, state, verifier string) error
	GetAndDeletePKCEVerifier(ctx context.Context, state string) (string, error)

	// Health check
	Health(ctx context.Context) error

	// Cleanup
	Close() error
}
