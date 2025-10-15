package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// StatePayload represents the encoded state parameter
type StatePayload struct {
	UserID     string `json:"uid"`
	ServerName string `json:"server"`
	Nonce      string `json:"n"`
}

// OAuthService handles OAuth operations
type OAuthService struct {
	config *oauth2.Config
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(config *oauth2.Config) *OAuthService {
	return &OAuthService{config: config}
}

// RandStringURLSafe generates a cryptographically secure random string
func RandStringURLSafe(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}

// PKCEChallenge generates a PKCE code challenge from a verifier
func PKCEChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// EncodeState encodes user ID, server name and nonce into a state parameter
func EncodeState(userID, serverName string) (string, error) {
	nonce, err := RandStringURLSafe(12)
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	payload := StatePayload{
		UserID:     userID,
		ServerName: serverName,
		Nonce:      nonce,
	}

	b, err := json.Marshal(&payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state payload: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DecodeState decodes a state parameter into user ID and nonce
func DecodeState(raw string) (*StatePayload, error) {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state: %w", err)
	}

	var payload StatePayload
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state payload: %w", err)
	}

	return &payload, nil
}

// GenerateAuthURL generates an OAuth authorization URL
func (s *OAuthService) GenerateAuthURL(state string, usePKCE bool) (string, string, error) {
	var authURL string
	var verifier string

	if usePKCE {
		// Generate PKCE verifier and challenge
		var err error
		verifier, err = RandStringURLSafe(64)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
		}

		codeChallenge := PKCEChallenge(verifier)
		authURL = s.config.AuthCodeURL(state,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
			oauth2.SetAuthURLParam("response_type", "code"),
			oauth2.SetAuthURLParam("scope", strings.Join(s.config.Scopes, " ")),
		)
	} else {
		// Standard OAuth flow with offline access for refresh tokens
		// For Google OAuth (YouTube), we need prompt=consent to ensure refresh token is returned
		if s.config.Endpoint.AuthURL == "https://accounts.google.com/o/oauth2/auth" {
			authURL = s.config.AuthCodeURL(state,
				oauth2.AccessTypeOffline,
				oauth2.SetAuthURLParam("prompt", "consent"),
			)
		} else {
			authURL = s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
		}
	}

	return authURL, verifier, nil
}

// ExchangeCode exchanges authorization code for access token
func (s *OAuthService) ExchangeCode(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	fmt.Printf("DEBUG: Starting token exchange\n")
	fmt.Printf("DEBUG: Code: %s\n", code)
	fmt.Printf("DEBUG: Verifier: %s (length: %d)\n", verifier, len(verifier))
	fmt.Printf("DEBUG: Token URL: %s\n", s.config.Endpoint.TokenURL)
	fmt.Printf("DEBUG: Client ID: %s\n", s.config.ClientID)

	var token *oauth2.Token
	var err error

	if verifier != "" {
		// PKCE flow - X platform requires special handling
		fmt.Printf("DEBUG: Using PKCE flow\n")

		// For X platform, we need to use a custom token exchange
		if s.config.Endpoint.TokenURL == "https://api.x.com/2/oauth2/token" {
			fmt.Printf("DEBUG: Using custom X platform token exchange\n")
			token, err = s.exchangeCodeWithPKCE(ctx, code, verifier)
		} else {
			token, err = s.config.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
		}
	} else {
		// Standard flow
		fmt.Printf("DEBUG: Using standard flow\n")
		token, err = s.config.Exchange(ctx, code)
	}

	// For Instagram, we need to exchange short-lived token for long-lived token
	if err == nil && s.config.Endpoint.TokenURL == "https://api.instagram.com/oauth/access_token" {
		fmt.Printf("DEBUG: Instagram detected, exchanging short-lived token for long-lived token\n")
		longLivedToken, exchangeErr := s.exchangeInstagramToken(ctx, token.AccessToken)
		if exchangeErr != nil {
			fmt.Printf("DEBUG: Instagram token exchange failed: %v\n", exchangeErr)
			// Continue with short-lived token if exchange fails
		} else {
			fmt.Printf("DEBUG: Instagram token exchange successful\n")
			token = longLivedToken
		}
	}

	// For Facebook, we need to exchange short-lived token for long-lived token
	if err == nil && s.config.Endpoint.TokenURL == "https://graph.facebook.com/v18.0/oauth/access_token" {
		fmt.Printf("DEBUG: Facebook detected, exchanging short-lived token for long-lived token\n")
		longLivedToken, exchangeErr := s.exchangeFacebookToken(ctx, token.AccessToken)
		if exchangeErr != nil {
			fmt.Printf("DEBUG: Facebook token exchange failed: %v\n", exchangeErr)
			// Continue with short-lived token if exchange fails
		} else {
			fmt.Printf("DEBUG: Facebook token exchange successful\n")
			token = longLivedToken
		}
	}

	if err != nil {
		fmt.Printf("DEBUG: Token exchange failed: %v\n", err)
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	fmt.Printf("DEBUG: Token exchange successful\n")
	fmt.Printf("DEBUG: Access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// exchangeCodeWithPKCE performs custom token exchange for X platform
func (s *OAuthService) exchangeCodeWithPKCE(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	fmt.Printf("DEBUG: Custom X platform token exchange\n")
	fmt.Printf("DEBUG: Code: %s\n", code)
	fmt.Printf("DEBUG: Verifier: %s\n", verifier)

	// Prepare the request data
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", s.config.ClientID)
	data.Set("redirect_uri", s.config.RedirectURL)
	data.Set("code_verifier", verifier)

	fmt.Printf("DEBUG: Request data: %s\n", data.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Add basic auth for client credentials
	auth := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	fmt.Printf("DEBUG: Request URL: %s\n", req.URL.String())
	fmt.Printf("DEBUG: Request headers: %v\n", req.Header)

	// Send the request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Create oauth2.Token
	token := &oauth2.Token{
		AccessToken:  tokenResponse.AccessToken,
		TokenType:    tokenResponse.TokenType,
		RefreshToken: tokenResponse.RefreshToken,
	}

	if tokenResponse.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	fmt.Printf("DEBUG: Token exchange successful\n")
	fmt.Printf("DEBUG: Access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// exchangeInstagramToken exchanges short-lived Instagram token for long-lived token
func (s *OAuthService) exchangeInstagramToken(ctx context.Context, shortLivedToken string) (*oauth2.Token, error) {
	fmt.Printf("DEBUG: Exchanging Instagram short-lived token for long-lived token\n")
	fmt.Printf("DEBUG: Short-lived token: %s\n", shortLivedToken)

	// Instagram uses a different endpoint for token exchange
	// According to Instagram API docs: https://graph.instagram.com/access_token
	exchangeURL := "https://graph.instagram.com/access_token"

	// Prepare the request data
	data := url.Values{}
	data.Set("grant_type", "ig_exchange_token")
	data.Set("client_secret", s.config.ClientSecret)
	data.Set("access_token", shortLivedToken)

	fmt.Printf("DEBUG: Request data: %s\n", data.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", exchangeURL+"?"+data.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	fmt.Printf("DEBUG: Request URL: %s\n", req.URL.String())
	fmt.Printf("DEBUG: Request headers: %v\n", req.Header)

	// Send the request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Create oauth2.Token
	token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		// For Instagram, the long-lived token itself can be used for refresh
		RefreshToken: tokenResponse.AccessToken,
	}

	if tokenResponse.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	fmt.Printf("DEBUG: Instagram token exchange successful\n")
	fmt.Printf("DEBUG: Long-lived access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// exchangeFacebookToken exchanges short-lived Facebook token for long-lived token
func (s *OAuthService) exchangeFacebookToken(ctx context.Context, shortLivedToken string) (*oauth2.Token, error) {
	fmt.Printf("DEBUG: Exchanging Facebook short-lived token for long-lived token\n")
	fmt.Printf("DEBUG: Short-lived token: %s\n", shortLivedToken)

	// Facebook uses a different endpoint for token exchange
	// According to Facebook API docs: https://graph.facebook.com/oauth/access_token
	exchangeURL := "https://graph.facebook.com/oauth/access_token"

	// Prepare the request data
	data := url.Values{}
	data.Set("grant_type", "fb_exchange_token")
	data.Set("client_id", s.config.ClientID)
	data.Set("client_secret", s.config.ClientSecret)
	data.Set("fb_exchange_token", shortLivedToken)

	fmt.Printf("DEBUG: Request data: %s\n", data.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", exchangeURL+"?"+data.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	fmt.Printf("DEBUG: Request URL: %s\n", req.URL.String())
	fmt.Printf("DEBUG: Request headers: %v\n", req.Header)

	// Send the request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Create oauth2.Token
	token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		// For Facebook, the long-lived token itself can be used for refresh
		RefreshToken: tokenResponse.AccessToken,
	}

	if tokenResponse.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	fmt.Printf("DEBUG: Facebook token exchange successful\n")
	fmt.Printf("DEBUG: Long-lived access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// RefreshToken refreshes an access token using refresh token
func (s *OAuthService) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	fmt.Printf("DEBUG: Starting token refresh\n")
	fmt.Printf("DEBUG: Refresh token: %s\n", refreshToken)
	fmt.Printf("DEBUG: Token URL: %s\n", s.config.Endpoint.TokenURL)
	fmt.Printf("DEBUG: Client ID: %s\n", s.config.ClientID)

	// For X platform, we need to use a custom refresh token exchange
	if s.config.Endpoint.TokenURL == "https://api.x.com/2/oauth2/token" {
		fmt.Printf("DEBUG: Using custom X platform token refresh\n")
		return s.refreshTokenWithX(ctx, refreshToken)
	}

	// For Instagram platform, we need to use Instagram-specific refresh endpoint
	if s.config.Endpoint.TokenURL == "https://api.instagram.com/oauth/access_token" {
		fmt.Printf("DEBUG: Using Instagram platform token refresh\n")
		return s.refreshTokenWithInstagram(ctx, refreshToken)
	}

	// For Facebook platform, we need to use Facebook-specific refresh endpoint
	if s.config.Endpoint.TokenURL == "https://graph.facebook.com/v18.0/oauth/access_token" {
		fmt.Printf("DEBUG: Using Facebook platform token refresh\n")
		return s.refreshTokenWithFacebook(ctx, refreshToken)
	}

	// For other platforms, use standard OAuth2 refresh
	fmt.Printf("DEBUG: Using standard OAuth2 token refresh\n")
	token, err := s.config.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		fmt.Printf("DEBUG: Token refresh failed: %v\n", err)
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	fmt.Printf("DEBUG: Token refresh successful\n")
	fmt.Printf("DEBUG: New access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// refreshTokenWithX performs custom token refresh for X platform
func (s *OAuthService) refreshTokenWithX(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	fmt.Printf("DEBUG: Custom X platform token refresh\n")
	fmt.Printf("DEBUG: Refresh token: %s\n", refreshToken)

	// Prepare the request data
	data := url.Values{}
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", s.config.ClientID)

	fmt.Printf("DEBUG: Request data: %s\n", data.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.Endpoint.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Add basic auth for client credentials
	auth := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	fmt.Printf("DEBUG: Request URL: %s\n", req.URL.String())
	fmt.Printf("DEBUG: Request headers: %v\n", req.Header)

	// Send the request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Create oauth2.Token
	token := &oauth2.Token{
		AccessToken:  tokenResponse.AccessToken,
		TokenType:    tokenResponse.TokenType,
		RefreshToken: tokenResponse.RefreshToken,
	}

	if tokenResponse.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	fmt.Printf("DEBUG: Token refresh successful\n")
	fmt.Printf("DEBUG: New access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// refreshTokenWithInstagram performs custom token refresh for Instagram platform
func (s *OAuthService) refreshTokenWithInstagram(ctx context.Context, accessToken string) (*oauth2.Token, error) {
	fmt.Printf("DEBUG: Custom Instagram platform token refresh\n")
	fmt.Printf("DEBUG: Access token: %s\n", accessToken)

	// Instagram uses a different refresh endpoint and parameters
	// According to Instagram API docs: https://graph.instagram.com/refresh_access_token
	refreshURL := "https://graph.instagram.com/refresh_access_token"

	// Prepare the request data
	data := url.Values{}
	data.Set("grant_type", "ig_refresh_token")
	data.Set("access_token", accessToken)

	fmt.Printf("DEBUG: Request data: %s\n", data.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", refreshURL+"?"+data.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	fmt.Printf("DEBUG: Request URL: %s\n", req.URL.String())
	fmt.Printf("DEBUG: Request headers: %v\n", req.Header)

	// Send the request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Create oauth2.Token
	token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		// Instagram doesn't provide refresh token in refresh response
		// The new access token becomes the new long-lived token
	}

	if tokenResponse.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	fmt.Printf("DEBUG: Instagram token refresh successful\n")
	fmt.Printf("DEBUG: New access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// refreshTokenWithFacebook performs custom token refresh for Facebook platform
func (s *OAuthService) refreshTokenWithFacebook(ctx context.Context, accessToken string) (*oauth2.Token, error) {
	fmt.Printf("DEBUG: Custom Facebook platform token refresh\n")
	fmt.Printf("DEBUG: Access token: %s\n", accessToken)

	// Facebook uses the same endpoint for token exchange and refresh
	// According to Facebook API docs: https://graph.facebook.com/oauth/access_token
	refreshURL := "https://graph.facebook.com/oauth/access_token"

	// Prepare the request data
	data := url.Values{}
	data.Set("grant_type", "fb_exchange_token")
	data.Set("client_id", s.config.ClientID)
	data.Set("client_secret", s.config.ClientSecret)
	data.Set("fb_exchange_token", accessToken)

	fmt.Printf("DEBUG: Request data: %s\n", data.Encode())

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", refreshURL+"?"+data.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")

	fmt.Printf("DEBUG: Request URL: %s\n", req.URL.String())
	fmt.Printf("DEBUG: Request headers: %v\n", req.Header)

	// Send the request
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Create oauth2.Token
	token := &oauth2.Token{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		// For Facebook, the long-lived token itself can be used for refresh
		RefreshToken: tokenResponse.AccessToken,
	}

	if tokenResponse.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	}

	fmt.Printf("DEBUG: Facebook token refresh successful\n")
	fmt.Printf("DEBUG: New access token: %s\n", token.AccessToken)
	fmt.Printf("DEBUG: Token type: %s\n", token.TokenType)
	fmt.Printf("DEBUG: Expiry: %v\n", token.Expiry)

	return token, nil
}

// CreateClient creates an HTTP client with automatic token refresh
func (s *OAuthService) CreateClient(ctx context.Context, token *oauth2.Token) *http.Client {
	ts := s.config.TokenSource(ctx, token)
	return oauth2.NewClient(ctx, ts)
}
