package platforms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
func (x *XPlatform) Share(ctx context.Context, client *http.Client, req *types.ShareRequest) (string, error) {
	if strings.TrimSpace(req.Content) == "" {
		return "", fmt.Errorf("content required for x/tweet")
	}

	type tweetReq struct {
		Text string `json:"text"`
	}

	payload := tweetReq{Text: req.Content}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tweet request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.x.com/2/tweets", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Parse response to get tweet ID
		var tweetResponse struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &tweetResponse); err == nil && tweetResponse.Data.ID != "" {
			return tweetResponse.Data.ID, nil
		}

		// Success but no ID returned
		return "", nil
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
				return "", fmt.Errorf("account suspended: %s", errorResponse.Detail)
			}
			return "", fmt.Errorf("access forbidden: %s", errorResponse.Detail)
		case 401:
			return "", fmt.Errorf("authentication failed: %s", errorResponse.Detail)
		case 429:
			return "", fmt.Errorf("rate limit exceeded: %s", errorResponse.Detail)
		default:
			return "", fmt.Errorf("x api error (%d): %s", errorResponse.Status, errorResponse.Detail)
		}
	}

	return "", fmt.Errorf("x api error: status=%d body=%s", resp.StatusCode, string(body))
}

// GetStats retrieves statistics from X (Twitter)
func (x *XPlatform) GetStats(ctx context.Context, client *http.Client, mediaID string) (types.StatsData, error) {
	if mediaID == "" {
		return types.StatsData{}, fmt.Errorf("media_id required")
	}

	url := fmt.Sprintf("https://api.x.com/2/tweets/%s?tweet.fields=public_metrics", mediaID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return types.StatsData{}, fmt.Errorf("x stats api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			PublicMetrics struct {
				RetweetCount int `json:"retweet_count"`
				LikeCount    int `json:"like_count"`
				ReplyCount   int `json:"reply_count"`
				QuoteCount   int `json:"quote_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return types.StatsData{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return types.StatsData{
		Likes:    result.Data.PublicMetrics.LikeCount,
		Retweets: result.Data.PublicMetrics.RetweetCount,
		Replies:  result.Data.PublicMetrics.ReplyCount,
		Shares:   result.Data.PublicMetrics.QuoteCount,
	}, nil
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

// GetUserInfo retrieves user information from X platform
func (x *XPlatform) GetUserInfo(ctx context.Context, client *http.Client) (types.UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.x.com/2/users/me?user.fields=id,username,name,email,profile_image_url,verified,public_metrics", nil)
	if err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to create user info request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to get user info: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to read user info response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse error response
		var errorResponse struct {
			Detail string `json:"detail"`
			Title  string `json:"title"`
			Status int    `json:"status"`
			Type   string `json:"type"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return types.UserInfo{}, fmt.Errorf("x user info api error (%d): %s", errorResponse.Status, errorResponse.Detail)
		}

		return types.UserInfo{}, fmt.Errorf("x user info api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var userResponse struct {
		Data struct {
			ID              string `json:"id"`
			Username        string `json:"username"`
			Name            string `json:"name"`
			Email           string `json:"email,omitempty"`
			ProfileImageURL string `json:"profile_image_url,omitempty"`
			Verified        bool   `json:"verified"`
			PublicMetrics   struct {
				FollowersCount int `json:"followers_count"`
				FollowingCount int `json:"following_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &userResponse); err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to parse user info response: %w", err)
	}

	// Build profile URL
	profileURL := fmt.Sprintf("https://x.com/%s", userResponse.Data.Username)

	return types.UserInfo{
		ID:          userResponse.Data.ID,
		Username:    userResponse.Data.Username,
		DisplayName: userResponse.Data.Name,
		Email:       userResponse.Data.Email,
		AvatarURL:   userResponse.Data.ProfileImageURL,
		ProfileURL:  profileURL,
		Verified:    userResponse.Data.Verified,
		Followers:   userResponse.Data.PublicMetrics.FollowersCount,
		Following:   userResponse.Data.PublicMetrics.FollowingCount,
	}, nil
}

// GetRecentPosts retrieves recent posts from X (Twitter)
func (x *XPlatform) GetRecentPosts(ctx context.Context, client *http.Client, limit int, startTime, endTime int64) ([]types.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// First, get the user ID
	userInfo, err := x.GetUserInfo(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Build query parameters
	params := fmt.Sprintf("max_results=%d&tweet.fields=id,text,created_at,public_metrics,attachments", limit)

	// Add time range filters if provided
	if startTime > 0 {
		// Handle both second and millisecond timestamps
		var startTimeUnix int64
		if startTime > 1e12 { // If timestamp is larger than 1e12, it's likely in milliseconds
			startTimeUnix = startTime / 1000
		} else {
			startTimeUnix = startTime
		}
		startTimeStr := time.Unix(startTimeUnix, 0).Format(time.RFC3339)
		params += fmt.Sprintf("&start_time=%s", startTimeStr)
	}
	if endTime > 0 {
		// Handle both second and millisecond timestamps
		var endTimeUnix int64
		if endTime > 1e12 { // If timestamp is larger than 1e12, it's likely in milliseconds
			endTimeUnix = endTime / 1000
		} else {
			endTimeUnix = endTime
		}
		endTimeStr := time.Unix(endTimeUnix, 0).Format(time.RFC3339)
		params += fmt.Sprintf("&end_time=%s", endTimeStr)
	}

	// Use the correct endpoint with user ID
	url := fmt.Sprintf("https://api.x.com/2/users/%s/tweets?%s", userInfo.ID, params)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse error response
		var errorResponse struct {
			Detail string `json:"detail"`
			Title  string `json:"title"`
			Status int    `json:"status"`
			Type   string `json:"type"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return nil, fmt.Errorf("x api error (%d): %s", errorResponse.Status, errorResponse.Detail)
		}

		return nil, fmt.Errorf("x api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var tweetsResponse struct {
		Data []struct {
			ID            string `json:"id"`
			Text          string `json:"text"`
			CreatedAt     string `json:"created_at"`
			PublicMetrics struct {
				RetweetCount int `json:"retweet_count"`
				LikeCount    int `json:"like_count"`
				ReplyCount   int `json:"reply_count"`
				QuoteCount   int `json:"quote_count"`
			} `json:"public_metrics"`
			Attachments struct {
				MediaKeys []string `json:"media_keys"`
			} `json:"attachments,omitempty"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &tweetsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse tweets response: %w", err)
	}

	// Convert to Post structs
	var posts []types.Post
	for _, tweet := range tweetsResponse.Data {
		// Parse created time
		createdTime, err := time.Parse(time.RFC3339, tweet.CreatedAt)
		if err != nil {
			createdTime = time.Now()
		}

		// Build tweet URL
		tweetURL := fmt.Sprintf("https://x.com/i/web/status/%s", tweet.ID)

		// Determine media type
		mediaType := ""
		if len(tweet.Attachments.MediaKeys) > 0 {
			mediaType = "image" // Default to image, could be enhanced to detect video
		}

		// Extract hashtags from tweet text
		tags := extractHashtags(tweet.Text)
		fmt.Printf("DEBUG: Tweet %s has %d tags: %v\n", tweet.ID, len(tags), tags)

		post := types.Post{
			ID:        tweet.ID,
			Content:   tweet.Text,
			CreatedAt: createdTime.Unix(),
			UpdatedAt: createdTime.Unix(), // X doesn't provide separate updated time
			Stats: types.StatsData{
				Likes:    tweet.PublicMetrics.LikeCount,
				Retweets: tweet.PublicMetrics.RetweetCount,
				Replies:  tweet.PublicMetrics.ReplyCount,
				Shares:   tweet.PublicMetrics.QuoteCount,
			},
			URL:       tweetURL,
			MediaType: mediaType,
			Tags:      tags,
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// extractHashtags extracts hashtags from tweet text
func extractHashtags(text string) []string {
	var hashtags []string
	words := strings.Fields(text)

	for _, word := range words {
		if strings.HasPrefix(word, "#") && len(word) > 1 {
			// Remove the # symbol and add to tags
			tag := strings.TrimPrefix(word, "#")
			// Remove any punctuation at the end
			tag = strings.TrimRight(tag, ".,!?;:")
			if tag != "" {
				hashtags = append(hashtags, tag)
			}
		}
	}

	return hashtags
}

// HandleOAuthCallback handles OAuth callback for X platform
func (x *XPlatform) HandleOAuthCallback(ctx context.Context, code, state string) error {
	// X平台特定的OAuth回调处理逻辑
	// 这里可以添加X平台特有的处理逻辑，比如特殊的PKCE验证等
	return nil
}
