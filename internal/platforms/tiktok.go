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

// TikTokPlatform implements the TikTok platform
type TikTokPlatform struct{}

// NewTikTokPlatform creates a new TikTok platform instance
func NewTikTokPlatform() *TikTokPlatform {
	return &TikTokPlatform{}
}

// GetName returns the platform name
func (t *TikTokPlatform) GetName() string {
	return "tiktok"
}

// Share shares content to TikTok
func (t *TikTokPlatform) Share(ctx context.Context, client *http.Client, req *types.ShareRequest) (string, error) {
	// TikTok for Developers API requires video upload
	// This is a simplified implementation - in production you need proper video handling

	if req.MediaURL == "" {
		return "", fmt.Errorf("media_url is required for TikTok video posts")
	}

	// TikTok API requires a multi-step process:
	// 1. Initialize video upload
	// 2. Upload video data
	// 3. Publish video

	// Step 1: Initialize video upload
	initData := map[string]any{
		"source_info": map[string]any{
			"source":            "FILE_UPLOAD",
			"video_size":        0, // This would be the actual file size
			"chunk_size":        0, // This would be the chunk size for upload
			"total_chunk_count": 0, // This would be calculated based on file size
		},
		"post_info": map[string]any{
			"title":                    req.Title,
			"description":              req.Content,
			"privacy_level":            "MUTUAL_FOLLOW_FRIEND", // Default privacy level
			"disable_duet":             false,
			"disable_comment":          false,
			"disable_stitch":           false,
			"video_cover_timestamp_ms": 1000,
		},
	}

	jsonData, err := json.Marshal(initData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tiktok init request: %w", err)
	}

	// Initialize upload
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://open-api.tiktok.com/share/video/upload/", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create tiktok init request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send tiktok init request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read tiktok init response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse error response
		var errorResponse struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				LogID   string `json:"log_id"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return "", fmt.Errorf("tiktok init api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return "", fmt.Errorf("tiktok init api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse init response
	var initResponse struct {
		Data struct {
			UploadURL string `json:"upload_url"`
			PublishID string `json:"publish_id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &initResponse); err != nil {
		return "", fmt.Errorf("failed to parse tiktok init response: %w", err)
	}

	// Note: In a real implementation, you would:
	// 1. Download the video from req.MediaURL
	// 2. Upload it to initResponse.Data.UploadURL
	// 3. Call the publish endpoint with initResponse.Data.PublishID

	// For now, we'll return the publish_id as a placeholder
	// In production, you need to complete the upload and publish process
	return initResponse.Data.PublishID, nil
}

// GetStats retrieves statistics from TikTok
func (t *TikTokPlatform) GetStats(ctx context.Context, client *http.Client, mediaID string) (types.StatsData, error) {
	if mediaID == "" {
		return types.StatsData{}, fmt.Errorf("media_id required")
	}

	// Get TikTok video statistics from TikTok for Developers API
	// Note: This requires proper authentication and may have limited data availability
	url := fmt.Sprintf("https://open-api.tiktok.com/video/query/?video_id=%s&fields=id,title,cover_image_url,embed_url,like_count,comment_count,share_count,view_count", mediaID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to create tiktok stats request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to get tiktok stats: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.StatsData{}, fmt.Errorf("failed to read tiktok stats response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Parse error response
		var errorResponse struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				LogID   string `json:"log_id"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return types.StatsData{}, fmt.Errorf("tiktok stats api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return types.StatsData{}, fmt.Errorf("tiktok stats api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var statsResponse struct {
		Data struct {
			Videos []struct {
				ID           string `json:"id"`
				Title        string `json:"title"`
				LikeCount    int    `json:"like_count"`
				CommentCount int    `json:"comment_count"`
				ShareCount   int    `json:"share_count"`
				ViewCount    int    `json:"view_count"`
			} `json:"videos"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &statsResponse); err != nil {
		return types.StatsData{}, fmt.Errorf("failed to parse tiktok stats response: %w", err)
	}

	if len(statsResponse.Data.Videos) == 0 {
		return types.StatsData{}, fmt.Errorf("video not found")
	}

	video := statsResponse.Data.Videos[0]

	return types.StatsData{
		Views:    video.ViewCount,
		Likes:    video.LikeCount,
		Replies:  video.CommentCount,
		Shares:   video.ShareCount,
		Retweets: 0, // TikTok doesn't have retweets
	}, nil
}

// GetUserInfo retrieves user information from TikTok platform
func (t *TikTokPlatform) GetUserInfo(ctx context.Context, client *http.Client) (types.UserInfo, error) {
	// TikTok for Developers API endpoint for user info
	// Note: TikTok API requires specific permissions and app approval
	req, err := http.NewRequestWithContext(ctx, "GET", "https://open-api.tiktok.com/user/info/?fields=open_id,union_id,avatar_url,display_name,follower_count,following_count,likes_count,video_count", nil)
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
		// Parse TikTok error response
		var errorResponse struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				LogID   string `json:"log_id"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			return types.UserInfo{}, fmt.Errorf("tiktok user info api error (%d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
		}

		return types.UserInfo{}, fmt.Errorf("tiktok user info api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var userResponse struct {
		Data struct {
			User struct {
				OpenID         string `json:"open_id"`
				UnionID        string `json:"union_id"`
				AvatarURL      string `json:"avatar_url"`
				DisplayName    string `json:"display_name"`
				FollowerCount  int    `json:"follower_count"`
				FollowingCount int    `json:"following_count"`
				LikesCount     int    `json:"likes_count"`
				VideoCount     int    `json:"video_count"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &userResponse); err != nil {
		return types.UserInfo{}, fmt.Errorf("failed to parse user info response: %w", err)
	}

	user := userResponse.Data.User

	// Build profile URL (TikTok doesn't provide direct profile URLs in API)
	profileURL := fmt.Sprintf("https://www.tiktok.com/@%s", user.OpenID)

	return types.UserInfo{
		ID:          user.OpenID,
		Username:    user.OpenID, // TikTok uses open_id as username
		DisplayName: user.DisplayName,
		Email:       "", // TikTok doesn't provide email in user info
		AvatarURL:   user.AvatarURL,
		ProfileURL:  profileURL,
		Verified:    false, // TikTok verification status is not available in basic user info
		Followers:   user.FollowerCount,
		Following:   user.FollowingCount,
	}, nil
}

// GetRecentPosts retrieves recent posts from TikTok
func (t *TikTokPlatform) GetRecentPosts(ctx context.Context, client *http.Client, limit int, startTime, endTime int64) ([]types.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Build query parameters
	params := fmt.Sprintf("max_count=%d&fields=id,create_time,share_url,title,cover_image_url,embed_url,like_count,comment_count,share_count", limit)

	// Add time range filters if provided
	if startTime > 0 {
		params += fmt.Sprintf("&start_time=%d", startTime)
	}
	if endTime > 0 {
		params += fmt.Sprintf("&end_time=%d", endTime)
	}

	url := fmt.Sprintf("https://open-api.tiktok.com/v2/user/info/?%s", params)
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
		return nil, fmt.Errorf("tiktok api error: status=%d body=%s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var videosResponse struct {
		Data struct {
			Videos []struct {
				ID            string `json:"id"`
				CreateTime    int64  `json:"create_time"`
				ShareURL      string `json:"share_url"`
				Title         string `json:"title"`
				CoverImageURL string `json:"cover_image_url"`
				EmbedURL      string `json:"embed_url"`
				LikeCount     int    `json:"like_count"`
				CommentCount  int    `json:"comment_count"`
				ShareCount    int    `json:"share_count"`
			} `json:"videos"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &videosResponse); err != nil {
		return nil, fmt.Errorf("failed to parse tiktok videos response: %w", err)
	}

	// Convert to Post structs
	var posts []types.Post
	for _, video := range videosResponse.Data.Videos {
		post := types.Post{
			ID:        video.ID,
			Content:   video.Title,
			Title:     video.Title,
			CreatedAt: video.CreateTime,
			Stats: types.StatsData{
				Likes:    video.LikeCount,
				Replies:  video.CommentCount,
				Shares:   video.ShareCount,
				Retweets: 0, // TikTok doesn't have retweets
			},
			URL:       video.ShareURL,
			MediaType: "video",
			MediaURL:  video.CoverImageURL,
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// HandleOAuthCallback handles OAuth callback for TikTok platform
func (t *TikTokPlatform) HandleOAuthCallback(ctx context.Context, code, state string) error {
	// TikTok平台特定的OAuth回调处理逻辑
	// 这里可以添加TikTok平台特有的处理逻辑

	// 验证授权码
	if code == "" {
		return fmt.Errorf("tiktok: authorization code is empty")
	}

	// 验证状态参数
	if state == "" {
		return fmt.Errorf("tiktok: state parameter is empty")
	}

	// 可以在这里添加额外的TikTok特定验证逻辑
	// 例如：验证用户权限、记录授权日志等

	return nil
}
