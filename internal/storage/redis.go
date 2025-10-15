package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

// RedisStorage implements token and PKCE storage using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: rdb}, nil
}

// TokenKey generates a Redis key for storing tokens
func (r *RedisStorage) TokenKey(userID, provider, serverName string) string {
	if serverName == "" {
		serverName = "default"
	}
	return fmt.Sprintf("token:%s:%s:%s", serverName, provider, userID)
}

// PKCEKey generates a Redis key for storing PKCE verifiers
func (r *RedisStorage) PKCEKey(state string) string {
	return fmt.Sprintf("pkce:%s", state)
}

// SaveToken stores an OAuth token in Redis with expiration
func (r *RedisStorage) SaveToken(ctx context.Context, userID, provider, serverName string, token *oauth2.Token) error {
	key := r.TokenKey(userID, provider, serverName)

	// Serialize token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Set expiration to 30 days (tokens should be refreshed before this)
	expiration := 30 * 24 * time.Hour

	// Debug: log the key and data size
	fmt.Printf("DEBUG: Saving token to Redis with key: %s, data size: %d bytes\n", key, len(data))

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		fmt.Printf("DEBUG: Failed to save token to Redis: %v\n", err)
		return err
	}

	fmt.Printf("DEBUG: Token saved successfully to Redis with key: %s\n", key)
	return nil
}

// GetToken retrieves an OAuth token from Redis
func (r *RedisStorage) GetToken(ctx context.Context, userID, provider, serverName string) (*oauth2.Token, error) {
	key := r.TokenKey(userID, provider, serverName)

	fmt.Printf("DEBUG: Looking for token in Redis with key: %s\n", key)

	// Test Redis connection first
	if err := r.client.Ping(ctx).Err(); err != nil {
		fmt.Printf("DEBUG: Redis connection test failed: %v\n", err)
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			fmt.Printf("DEBUG: Token not found in Redis with key: %s\n", key)
			return nil, fmt.Errorf("token not found")
		}
		fmt.Printf("DEBUG: Failed to get token from Redis: %v\n", err)
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	fmt.Printf("DEBUG: Token found in Redis with key: %s, data size: %d bytes\n", key, len(data))

	var token oauth2.Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		fmt.Printf("DEBUG: Failed to unmarshal token: %v\n", err)
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	fmt.Printf("DEBUG: Token retrieved successfully from Redis with key: %s, access_token length: %d\n", key, len(token.AccessToken))
	return &token, nil
}

// DeleteToken removes an OAuth token from Redis
func (r *RedisStorage) DeleteToken(ctx context.Context, userID, provider, serverName string) error {
	key := r.TokenKey(userID, provider, serverName)
	return r.client.Del(ctx, key).Err()
}

// SavePKCEVerifier stores a PKCE verifier in Redis with short expiration
func (r *RedisStorage) SavePKCEVerifier(ctx context.Context, state, verifier string) error {
	key := r.PKCEKey(state)

	// PKCE verifiers should expire quickly (30 minutes to allow for user interaction time)
	expiration := 30 * time.Minute

	fmt.Printf("DEBUG: Saving PKCE verifier to Redis with key: %s, verifier length: %d\n", key, len(verifier))

	// Test Redis connection first
	if err := r.client.Ping(ctx).Err(); err != nil {
		fmt.Printf("DEBUG: Redis connection test failed: %v\n", err)
		return fmt.Errorf("redis connection failed: %w", err)
	}

	err := r.client.Set(ctx, key, verifier, expiration).Err()
	if err != nil {
		fmt.Printf("DEBUG: Failed to save PKCE verifier to Redis: %v\n", err)
		return err
	}

	// Verify the save was successful
	savedVerifier, err := r.client.Get(ctx, key).Result()
	if err != nil {
		fmt.Printf("DEBUG: Failed to verify PKCE verifier save: %v\n", err)
		return fmt.Errorf("failed to verify PKCE verifier save: %w", err)
	}

	if savedVerifier != verifier {
		fmt.Printf("DEBUG: PKCE verifier mismatch after save\n")
		return fmt.Errorf("PKCE verifier mismatch after save")
	}

	fmt.Printf("DEBUG: PKCE verifier saved and verified successfully to Redis with key: %s\n", key)
	return nil
}

// GetAndDeletePKCEVerifier retrieves and deletes a PKCE verifier from Redis
func (r *RedisStorage) GetAndDeletePKCEVerifier(ctx context.Context, state string) (string, error) {
	key := r.PKCEKey(state)

	fmt.Printf("DEBUG: Looking for PKCE verifier in Redis with key: %s\n", key)

	// Use Redis pipeline for atomic get and delete
	pipe := r.client.Pipeline()
	getCmd := pipe.Get(ctx, key)
	delCmd := pipe.Del(ctx, key)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		fmt.Printf("DEBUG: Failed to execute Redis pipeline for PKCE verifier: %v\n", err)
		return "", fmt.Errorf("failed to get PKCE verifier: %w", err)
	}

	verifier, err := getCmd.Result()
	if err != nil {
		if err == redis.Nil {
			fmt.Printf("DEBUG: PKCE verifier not found in Redis with key: %s\n", key)
			return "", fmt.Errorf("PKCE verifier not found or expired")
		}
		fmt.Printf("DEBUG: Failed to get PKCE verifier from Redis: %v\n", err)
		return "", fmt.Errorf("failed to get PKCE verifier: %w", err)
	}

	fmt.Printf("DEBUG: PKCE verifier found in Redis with key: %s, verifier length: %d\n", key, len(verifier))

	// Check if delete was successful
	if delCmd.Err() != nil {
		fmt.Printf("DEBUG: Failed to delete PKCE verifier from Redis: %v\n", delCmd.Err())
		return "", fmt.Errorf("failed to delete PKCE verifier: %w", delCmd.Err())
	}

	fmt.Printf("DEBUG: PKCE verifier retrieved and deleted successfully from Redis with key: %s\n", key)
	return verifier, nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}

// Health checks Redis connection health
func (r *RedisStorage) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
