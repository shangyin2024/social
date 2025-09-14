package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"social/internal/types"
)

// XPlatform implements the X (Twitter) platform
type XPlatform struct{}

// NewXPlatform creates a new X platform instance
func NewXPlatform() *XPlatform {
	return &XPlatform{}
}

// GetName returns the platform name
func (x *XPlatform) GetName() string {
	return "x"
}

// Share shares content to X (Twitter)
func (x *XPlatform) Share(ctx context.Context, client *http.Client, req *types.ShareRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return fmt.Errorf("content required for x/tweet")
	}

	type tweetReq struct {
		Text string `json:"text"`
	}

	payload := tweetReq{Text: req.Content}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.x.com/2/tweets", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Parse response to get tweet ID
		var tweetResponse struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		
		if err := json.Unmarshal(body, &tweetResponse); err == nil && tweetResponse.Data.ID != "" {
			// Store tweet ID in context for later use
			ctx = context.WithValue(ctx, "tweet_id", tweetResponse.Data.ID)
		}
		
		// Success
		return nil
	}

	// Parse error response for better error handling
	var errorResponse struct {
		Detail string `json:"detail"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Type   string `json:"type"`
	}
	
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		// Handle specific error cases
		switch errorResponse.Status {
		case 403:
			if strings.Contains(errorResponse.Detail, "suspended") {
				return fmt.Errorf("account suspended: %s", errorResponse.Detail)
			}
			return fmt.Errorf("access forbidden: %s", errorResponse.Detail)
		case 401:
			return fmt.Errorf("authentication failed: %s", errorResponse.Detail)
		case 429:
			return fmt.Errorf("rate limit exceeded: %s", errorResponse.Detail)
		default:
			return fmt.Errorf("x api error (%d): %s", errorResponse.Status, errorResponse.Detail)
		}
	}
	
	return fmt.Errorf("x api error: status=%d body=%s", resp.StatusCode, string(body))
}

// GetStats retrieves statistics from X (Twitter)
func (x *XPlatform) GetStats(ctx context.Context, client *http.Client, mediaID string) (map[string]interface{}, error) {
	if mediaID == "" {
		return nil, fmt.Errorf("media_id required")
	}

	url := fmt.Sprintf("https://api.x.com/2/tweets/%s?tweet.fields=public_metrics", mediaID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("x stats api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// CheckAccountStatus checks if the X account is in good standing
func (x *XPlatform) CheckAccountStatus(ctx context.Context, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.x.com/2/users/me", nil)
	if err != nil {
		return fmt.Errorf("failed to create account status request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check account status: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read account status response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Account is in good standing
		return nil
	}

	// Parse error response
	var errorResponse struct {
		Detail string `json:"detail"`
		Title  string `json:"title"`
		Status int    `json:"status"`
		Type   string `json:"type"`
	}
	
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		switch errorResponse.Status {
		case 403:
			if strings.Contains(errorResponse.Detail, "suspended") {
				return fmt.Errorf("account suspended: %s", errorResponse.Detail)
			}
			return fmt.Errorf("access forbidden: %s", errorResponse.Detail)
		case 401:
			return fmt.Errorf("authentication failed: %s", errorResponse.Detail)
		default:
			return fmt.Errorf("account status check failed (%d): %s", errorResponse.Status, errorResponse.Detail)
		}
	}
	
	return fmt.Errorf("account status check failed: status=%d body=%s", resp.StatusCode, string(body))
}

// HandleOAuthCallback handles OAuth callback for X platform
func (x *XPlatform) HandleOAuthCallback(ctx context.Context, code, state string) error {
	// X平台特定的OAuth回调处理逻辑
	// 这里可以添加X平台特有的处理逻辑，比如特殊的PKCE验证等
	return nil
}
