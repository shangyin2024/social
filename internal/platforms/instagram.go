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

// InstagramPlatform implements the Instagram platform
type InstagramPlatform struct{}

// NewInstagramPlatform creates a new Instagram platform instance
func NewInstagramPlatform() *InstagramPlatform {
	return &InstagramPlatform{}
}

// GetName returns the platform name
func (i *InstagramPlatform) GetName() string {
	return "instagram"
}

// Share shares content to Instagram
func (i *InstagramPlatform) Share(ctx context.Context, client *http.Client, req *types.ShareRequest) (string, error) {
	// Instagram Graph API requires Instagram Business Account connected to Facebook Page
	// This is a simplified implementation for photo posts
	// For production, you need proper media upload handling

	if req.MediaURL == "" {
		return "", fmt.Errorf("media_url is required for Instagram posts")
	}

	// Step 1: Create media container
	mediaData := map[string]any{
		"image_url": req.MediaURL,
		"caption":   req.Content,
	}

	jsonData, err := json.Marshal(mediaData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal instagram media request: %w", err)
	}

	// Create media container
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://graph.facebook.com/me/media", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create instagram media request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send instagram media request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read instagram media response: %w", err)
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
			return "", fmt.Errorf("instagram media api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return "", fmt.Errorf("instagram media api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse media container response
	var mediaResponse struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(body, &mediaResponse); err != nil {
		return "", fmt.Errorf("failed to parse instagram media response: %w", err)
	}

	if mediaResponse.ID == "" {
		return "", fmt.Errorf("no media container ID in response")
	}

	// Step 2: Publish the media container
	publishData := map[string]any{
		"creation_id": mediaResponse.ID,
	}

	publishJSON, err := json.Marshal(publishData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal instagram publish request: %w", err)
	}

	// Publish media
	publishReq, err := http.NewRequestWithContext(ctx, "POST", "https://graph.facebook.com/me/media_publish", strings.NewReader(string(publishJSON)))
	if err != nil {
		return "", fmt.Errorf("failed to create instagram publish request: %w", err)
	}

	publishReq.Header.Set("Content-Type", "application/json")

	publishResp, err := client.Do(publishReq)
	if err != nil {
		return "", fmt.Errorf("failed to send instagram publish request: %w", err)
	}
	defer func() {
		_ = publishResp.Body.Close()
	}()

	publishBody, err := io.ReadAll(publishResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read instagram publish response: %w", err)
	}

	if publishResp.StatusCode < 200 || publishResp.StatusCode >= 300 {
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

		if err := json.Unmarshal(publishBody, &errorResponse); err == nil {
			return "", fmt.Errorf("instagram publish api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return "", fmt.Errorf("instagram publish api error: status=%d body=%s", publishResp.StatusCode, string(publishBody))
	}

	// Parse publish response
	var publishResponse struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(publishBody, &publishResponse); err != nil {
		return "", fmt.Errorf("failed to parse instagram publish response: %w", err)
	}

	return publishResponse.ID, nil
}

// GetStats retrieves statistics from Instagram
func (i *InstagramPlatform) GetStats(ctx context.Context, client *http.Client, mediaID string) (types.StatsData, error) {
	if mediaID == "" {
		return types.StatsData{}, fmt.Errorf("media_id required")
	}

	// Get Instagram media insights from Graph API
	// Note: This requires Instagram Business Account and may have limited data availability
	url := fmt.Sprintf("https://graph.facebook.com/%s?fields=like_count,comments_count,media_type", mediaID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to create instagram stats request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to get instagram stats: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to read instagram stats response: %w", err)
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
			return types.StatsData{}, fmt.Errorf("instagram stats api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return types.StatsData{}, fmt.Errorf("instagram stats api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var statsResponse struct {
		LikeCount     int    `json:"like_count"`
		CommentsCount int    `json:"comments_count"`
		MediaType     string `json:"media_type"`
	}

	if err := json.Unmarshal(body, &statsResponse); err != nil {
		return types.StatsData{}, fmt.Errorf("failed to parse instagram stats response: %w", err)
	}

	return types.StatsData{
		Likes:    statsResponse.LikeCount,
		Replies:  statsResponse.CommentsCount,
		Shares:   0, // Instagram doesn't provide share count in basic stats
		Retweets: 0, // Instagram doesn't have retweets
		Views:    0, // Instagram doesn't provide view count in basic stats
	}, nil
}

// GetUserInfo retrieves user information from Instagram platform
func (i *InstagramPlatform) GetUserInfo(ctx context.Context, client *http.Client) (types.UserInfo, error) {
	// Instagram Graph API endpoint for user info
	// Note: Instagram requires Instagram Business Account connected to Facebook Page
	req, err := http.NewRequestWithContext(ctx, "GET", "https://graph.facebook.com/me?fields=id,name,username,profile_picture_url,biography,followers_count,follows_count,media_count", nil)
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
		// Parse Instagram/Facebook error response
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
			return types.UserInfo{}, fmt.Errorf("instagram user info api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return types.UserInfo{}, fmt.Errorf("instagram user info api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var userResponse struct {
		ID                string `json:"id"`
		Name              string `json:"name"`
		Username          string `json:"username"`
		ProfilePictureURL string `json:"profile_picture_url,omitempty"`
		Biography         string `json:"biography,omitempty"`
		FollowersCount    int    `json:"followers_count,omitempty"`
		FollowsCount      int    `json:"follows_count,omitempty"`
		MediaCount        int    `json:"media_count,omitempty"`
	}

	if err := json.Unmarshal(body, &userResponse); err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to parse user info response: %w", err)
	}

	// Build profile URL
	profileURL := fmt.Sprintf("https://www.instagram.com/%s", userResponse.Username)

	return types.UserInfo{
		ID:          userResponse.ID,
		Username:    userResponse.Username,
		DisplayName: userResponse.Name,
		Email:       "", // Instagram doesn't provide email in user info
		AvatarURL:   userResponse.ProfilePictureURL,
		ProfileURL:  profileURL,
		Verified:    false, // Instagram verification status is not available in basic user info
		Followers:   userResponse.FollowersCount,
		Following:   userResponse.FollowsCount,
	}, nil
}

// GetRecentPosts retrieves recent posts from Instagram
func (i *InstagramPlatform) GetRecentPosts(ctx context.Context, client *http.Client, limit int, startTime, endTime int64) ([]types.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Build query parameters
	params := fmt.Sprintf("limit=%d&fields=id,caption,media_type,media_url,permalink,thumbnail_url,timestamp,like_count,comments_count", limit)

	// Add time range filters if provided
	if startTime > 0 {
		startTimeStr := time.Unix(startTime, 0).Format(time.RFC3339)
		params += fmt.Sprintf("&since=%s", startTimeStr)
	}
	if endTime > 0 {
		endTimeStr := time.Unix(endTime, 0).Format(time.RFC3339)
		params += fmt.Sprintf("&until=%s", endTimeStr)
	}

	url := fmt.Sprintf("https://graph.instagram.com/me/media?%s", params)
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
		return nil, fmt.Errorf("instagram api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var mediaResponse struct {
		Data []struct {
			ID            string `json:"id"`
			Caption       string `json:"caption"`
			MediaType     string `json:"media_type"`
			MediaURL      string `json:"media_url"`
			Permalink     string `json:"permalink"`
			ThumbnailURL  string `json:"thumbnail_url"`
			Timestamp     string `json:"timestamp"`
			LikeCount     int    `json:"like_count"`
			CommentsCount int    `json:"comments_count"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &mediaResponse); err != nil {
		return nil, fmt.Errorf("failed to parse instagram media response: %w", err)
	}

	// Convert to Post structs
	var posts []types.Post
	for _, media := range mediaResponse.Data {
		// Parse timestamp
		timestamp, err := time.Parse(time.RFC3339, media.Timestamp)
		if err != nil {
			timestamp = time.Now()
		}

		// Determine media type
		mediaType := media.MediaType
		if mediaType == "" {
			mediaType = "image" // Default to image
		}

		// Use thumbnail URL if available, otherwise use media URL
		mediaURL := media.MediaURL
		if media.ThumbnailURL != "" {
			mediaURL = media.ThumbnailURL
		}

		post := types.Post{
			ID:        media.ID,
			Content:   media.Caption,
			CreatedAt: timestamp.Unix(),
			Stats: types.StatsData{
				Likes:    media.LikeCount,
				Replies:  media.CommentsCount,
				Shares:   0, // Instagram doesn't provide share count in basic API
				Retweets: 0, // Instagram doesn't have retweets
			},
			URL:       media.Permalink,
			MediaType: mediaType,
			MediaURL:  mediaURL,
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// HandleOAuthCallback handles OAuth callback for Instagram platform
func (i *InstagramPlatform) HandleOAuthCallback(ctx context.Context, code, state string) error {
	// Instagram平台特定的OAuth回调处理逻辑
	// 这里可以添加Instagram平台特有的处理逻辑
	return nil
}
