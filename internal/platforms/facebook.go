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

// FacebookPlatform implements the Facebook platform
type FacebookPlatform struct{}

// NewFacebookPlatform creates a new Facebook platform instance
func NewFacebookPlatform() *FacebookPlatform {
	return &FacebookPlatform{}
}

// GetName returns the platform name
func (f *FacebookPlatform) GetName() string {
	return "facebook"
}

// Share shares content to Facebook
func (f *FacebookPlatform) Share(ctx context.Context, client *http.Client, req *types.ShareRequest) (string, error) {
	// Facebook Graph API requires page access token and page ID
	// For now, we'll implement basic user feed posting
	// In production, you should use page access tokens for business accounts

	if strings.TrimSpace(req.Content) == "" {
		return "", fmt.Errorf("content required for facebook post")
	}

	// Prepare post data
	postData := map[string]any{
		"message": req.Content,
	}

	// Add media if provided
	if req.MediaURL != "" {
		// For media posts, we need to use a different approach
		// This is a simplified implementation - in production you'd need to handle media uploads properly
		postData["link"] = req.MediaURL
	}

	jsonData, err := json.Marshal(postData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal facebook post request: %w", err)
	}

	// Post to user's feed
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://graph.facebook.com/me/feed", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create facebook post request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send facebook post request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read facebook post response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Parse response to get post ID
		var postResponse struct {
			ID string `json:"id"`
		}

		if err := json.Unmarshal(body, &postResponse); err == nil && postResponse.ID != "" {
			return postResponse.ID, nil
		}

		// Success but no ID returned
		return "", nil
	}

	// Parse error response
	var errorResponse struct {
		Error struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
			SubCode   int    `json:"error_subcode,omitempty"`
			FBTraceID string `json:"fbtrace_id"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errorResponse); err == nil {
		return "", fmt.Errorf("facebook api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
	}

	return "", fmt.Errorf("facebook api error: status=%d body=%s", resp.StatusCode, string(body))
}

// GetStats retrieves statistics from Facebook
func (f *FacebookPlatform) GetStats(ctx context.Context, client *http.Client, mediaID string) (types.StatsData, error) {
	if mediaID == "" {
		return types.StatsData{}, fmt.Errorf("media_id required")
	}

	// Get post insights from Facebook Graph API
	// Note: This requires the post to be published and may have limited data availability
	url := fmt.Sprintf("https://graph.facebook.com/%s?fields=likes.summary(true),comments.summary(true),shares", mediaID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to create facebook stats request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to get facebook stats: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to read facebook stats response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse error response
		var errorResponse struct {
			Error struct {
				Message   string `json:"message"`
				Type      string `json:"type"`
				Code      int    `json:"code"`
				SubCode   int    `json:"error_subcode,omitempty"`
				FBTraceID string `json:"fbtrace_id"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return types.StatsData{}, fmt.Errorf("facebook stats api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return types.StatsData{}, fmt.Errorf("facebook stats api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var statsResponse struct {
		Likes struct {
			Summary struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		} `json:"likes"`
		Comments struct {
			Summary struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		} `json:"comments"`
		Shares struct {
			Count int `json:"count"`
		} `json:"shares"`
	}

	if err := json.Unmarshal(body, &statsResponse); err != nil {
		return types.StatsData{}, fmt.Errorf("failed to parse facebook stats response: %w", err)
	}

	return types.StatsData{
		Likes:    statsResponse.Likes.Summary.TotalCount,
		Replies:  statsResponse.Comments.Summary.TotalCount,
		Shares:   statsResponse.Shares.Count,
		Retweets: 0, // Facebook doesn't have retweets
	}, nil
}

// GetUserInfo retrieves user information from Facebook platform
func (f *FacebookPlatform) GetUserInfo(ctx context.Context, client *http.Client) (types.UserInfo, error) {
	// Facebook Graph API endpoint for user info
	// Note: Facebook requires specific permissions to access user info
	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.facebook.com/me?fields=id,name,email,picture,verified", nil)
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
		// Parse Facebook error response
		var errorResponse struct {
			Error struct {
				Message   string `json:"message"`
				Type      string `json:"type"`
				Code      int    `json:"code"`
				SubCode   int    `json:"error_subcode,omitempty"`
				FBTraceID string `json:"fbtrace_id"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return types.UserInfo{}, fmt.Errorf("facebook user info api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return types.UserInfo{}, fmt.Errorf("facebook user info api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var userResponse struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email,omitempty"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture,omitempty"`
		Verified bool `json:"verified,omitempty"`
	}

	if err := json.Unmarshal(body, &userResponse); err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to parse user info response: %w", err)
	}

	// Build profile URL
	profileURL := fmt.Sprintf("https://facebook.com/%s", userResponse.ID)

	return types.UserInfo{
		ID:          userResponse.ID,
		Username:    userResponse.ID, // Facebook uses ID as username
		DisplayName: userResponse.Name,
		Email:       userResponse.Email,
		AvatarURL:   userResponse.Picture.Data.URL,
		ProfileURL:  profileURL,
		Verified:    userResponse.Verified,
		// Facebook doesn't provide follower/following counts in basic user info
		Followers: 0,
		Following: 0,
	}, nil
}

// GetRecentPosts retrieves recent posts from Facebook
func (f *FacebookPlatform) GetRecentPosts(ctx context.Context, client *http.Client, limit int, startTime, endTime int64) ([]types.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Build query parameters
	params := fmt.Sprintf("limit=%d&fields=id,message,created_time,updated_time,likes.summary(true),comments.summary(true),shares", limit)

	// Add time range filters if provided
	if startTime > 0 {
		startTimeStr := time.Unix(startTime, 0).Format(time.RFC3339)
		params += fmt.Sprintf("&since=%s", startTimeStr)
	}
	if endTime > 0 {
		endTimeStr := time.Unix(endTime, 0).Format(time.RFC3339)
		params += fmt.Sprintf("&until=%s", endTimeStr)
	}

	url := fmt.Sprintf("https://graph.facebook.com/me/feed?%s", params)
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
		return nil, fmt.Errorf("facebook api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var postsResponse struct {
		Data []struct {
			ID          string `json:"id"`
			Message     string `json:"message"`
			CreatedTime string `json:"created_time"`
			UpdatedTime string `json:"updated_time,omitempty"`
			Likes       struct {
				Summary struct {
					TotalCount int `json:"total_count"`
				} `json:"summary"`
			} `json:"likes"`
			Comments struct {
				Summary struct {
					TotalCount int `json:"total_count"`
				} `json:"summary"`
			} `json:"comments"`
			Shares struct {
				Count int `json:"count"`
			} `json:"shares"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &postsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse facebook posts response: %w", err)
	}

	// Convert to Post structs
	var posts []types.Post
	for _, post := range postsResponse.Data {
		// Parse created time
		createdTime, err := time.Parse(time.RFC3339, post.CreatedTime)
		if err != nil {
			createdTime = time.Now()
		}

		// Parse updated time if available
		var updatedTime int64
		if post.UpdatedTime != "" {
			if parsed, err := time.Parse(time.RFC3339, post.UpdatedTime); err == nil {
				updatedTime = parsed.Unix()
			}
		}

		// Build post URL
		postURL := fmt.Sprintf("https://www.facebook.com/%s", post.ID)

		postData := types.Post{
			ID:        post.ID,
			Content:   post.Message,
			CreatedAt: createdTime.Unix(),
			UpdatedAt: updatedTime,
			Stats: types.StatsData{
				Likes:    post.Likes.Summary.TotalCount,
				Replies:  post.Comments.Summary.TotalCount,
				Shares:   post.Shares.Count,
				Retweets: 0, // Facebook doesn't have retweets
			},
			URL:       postURL,
			MediaType: "text", // Default to text, could be enhanced to detect media
		}

		posts = append(posts, postData)
	}

	return posts, nil
}

// HandleOAuthCallback handles OAuth callback for Facebook platform
func (f *FacebookPlatform) HandleOAuthCallback(ctx context.Context, code, state string) error {
	// Facebook平台特定的OAuth回调处理逻辑
	// 这里可以添加Facebook平台特有的处理逻辑
	return nil
}
