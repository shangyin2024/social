package types

import (
	"context"
	"net/http"
)

// ShareRequest represents a request to share content to a social platform
type ShareRequest struct {
	Provider   string   `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"`   // 平台名称 可选值：youtube x facebook tiktok instagram
	UserID     string   `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                          // 用户ID 必填 同一服务名称下user_id唯一
	ServerName string   `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`                         // 服务名称 必填
	Content    string   `json:"content,omitempty" binding:"max=280" example:"Hello World!"`                          // text content
	MediaURL   string   `json:"media_url,omitempty" binding:"omitempty,url" example:"https://example.com/image.jpg"` // url to media (backend should download & upload)
	Title      string   `json:"title,omitempty" binding:"max=100" example:"My Post"`
	Desc       string   `json:"description,omitempty" binding:"max=500" example:"This is a description"`
	Tags       []string `json:"tags,omitempty" binding:"max=10" example:"hello,world"`
	Privacy    string   `json:"privacy,omitempty" binding:"omitempty,oneof=public private unlisted friends followers" example:"public"`
}

// StatsRequest represents a request to get statistics from a social platform
type StatsRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称 可选值：youtube x facebook tiktok instagram
	UserID     string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                        // 用户ID 必填 同一服务名称下user_id唯一
	ServerName string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`
	MediaID    string `json:"media_id,omitempty" binding:"max=100" example:"1234567890"`
}

// StartAuthRequest represents a request to start OAuth authentication
type StartAuthRequest struct {
	Provider    string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称 可选值：youtube x facebook tiktok instagram
	UserID      string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                        // 用户ID 必填 同一服务名称下user_id唯一
	RedirectURI string `json:"redirect_uri" binding:"required,url" example:"https://test-pubproject.wondera.io/static/callback.html"`
	ServerName  string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`
}

// CallbackRequest represents a request for OAuth callback
// 前端收到OAuth回调后，调用此接口处理授权码交换
type CallbackRequest struct {
	Provider    string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"`                      // 平台名称 可选值：youtube x facebook tiktok instagram
	ServerName  string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`                                            // 服务器名称
	UserID      string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                                             // 服务内部用户ID 必填
	State       string `json:"state" binding:"required,min=1" example:"encoded_state_string"`                                          // 状态参数，包含用户ID等信息
	Code        string `json:"code" binding:"required,min=1" example:"authorization_code"`                                             // 授权码
	RedirectURI string `json:"redirect_uri" binding:"required,url" example:"hhttps://test-pubproject.wondera.io/static/callback.html"` // 重定向URI
}

// StartAuthResponse represents the response for OAuth authorization start
type StartAuthResponse struct {
	AuthURL    string `json:"auth_url" example:"https://x.com/i/oauth2/authorize?response_type=code&client_id=..."`
	Provider   string `json:"provider" example:"x"`
	UserID     string `json:"user_id" example:"user123"`
	ServerName string `json:"server_name" example:"myapp"`
}

// CallbackResponse represents the response for OAuth callback
type CallbackResponse struct {
	Provider   string `json:"provider" example:"x"`
	UserID     string `json:"user_id" example:"user123"`
	ServerName string `json:"server_name" example:"myapp"`
	ExpiresAt  int64  `json:"expires_at" example:"1704067199"` // 时间戳格式
	ReferAt    int64  `json:"refer_at" example:"1704067199"`   // 时间戳格式
	Message    string `json:"message" example:"OAuth callback completed successfully"`
}

// ShareResponse represents the response for content sharing
type ShareResponse struct {
	Provider   string   `json:"provider" example:"x"`
	UserID     string   `json:"user_id" example:"user123"`
	ServerName string   `json:"server_name" example:"myapp"`
	Content    string   `json:"content" example:"Hello from Social Platform! 🚀"`
	MediaURL   string   `json:"media_url,omitempty" example:"https://example.com/image.jpg"`
	Tags       []string `json:"tags,omitempty" example:"social,oauth,test"`
	MediaID    string   `json:"media_id,omitempty" example:"1234567890"` // Tweet ID or post ID for status query
}

// StatsData represents the statistics data structure
type StatsData struct {
	Likes    int `json:"likes" example:"100"`
	Retweets int `json:"retweets" example:"50"`
	Replies  int `json:"replies" example:"25"`
	Views    int `json:"views,omitempty" example:"1000"`
	Shares   int `json:"shares,omitempty" example:"10"`
}

// StatsResponse represents the response for statistics
type StatsResponse struct {
	Provider   string    `json:"provider" example:"x"`
	UserID     string    `json:"user_id" example:"user123"`
	ServerName string    `json:"server_name" example:"myapp"`
	MediaID    string    `json:"media_id" example:"1234567890"`
	Stats      StatsData `json:"stats"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// UserInfo represents user information from a social platform
type UserInfo struct {
	ID          string `json:"id" example:"1234567890"`                                       // 平台用户ID
	Username    string `json:"username" example:"johndoe"`                                    // 用户名
	DisplayName string `json:"display_name" example:"John Doe"`                               // 显示名称
	Email       string `json:"email,omitempty" example:"john@example.com"`                    // 邮箱（如果可用）
	AvatarURL   string `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"` // 头像URL
	ProfileURL  string `json:"profile_url,omitempty" example:"https://x.com/johndoe"`         // 个人资料URL
	Verified    bool   `json:"verified" example:"false"`                                      // 是否认证用户
	Followers   int    `json:"followers,omitempty" example:"1000"`                            // 粉丝数
	Following   int    `json:"following,omitempty" example:"500"`                             // 关注数
}

// GetUserInfoRequest represents a request to get user information
type GetUserInfoRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称
	UserID     string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                        // 用户ID
	ServerName string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`                       // 服务名称
}

// GetUserInfoResponse represents the response for user information
type GetUserInfoResponse struct {
	Provider   string   `json:"provider" example:"x"`
	UserID     string   `json:"user_id" example:"user123"`
	ServerName string   `json:"server_name" example:"myapp"`
	UserInfo   UserInfo `json:"user_info"`
}

// Platform represents a social media platform interface
type Platform interface {
	// Share shares content to the platform and returns the media ID
	Share(ctx context.Context, client *http.Client, req *ShareRequest) (string, error)

	// GetStats retrieves statistics from the platform
	GetStats(ctx context.Context, client *http.Client, mediaID string) (StatsData, error)

	// GetUserInfo retrieves user information from the platform
	GetUserInfo(ctx context.Context, client *http.Client) (UserInfo, error)

	// GetRecentPosts retrieves recent posts from the platform
	GetRecentPosts(ctx context.Context, client *http.Client, limit int, startTime, endTime int64) ([]Post, error)

	// GetName returns the platform name
	GetName() string

	// HandleOAuthCallback handles OAuth callback for the platform
	HandleOAuthCallback(ctx context.Context, code, state string) error
}

// IsAuthorizedRequest represents a request to check if a user is authorized for a platform
type IsAuthorizedRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"`
	UserID     string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`
	ServerName string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`
}

// IsAuthorizedResponse represents a response to check if a user is authorized for a platform
type IsAuthorizedResponse struct {
	IsAuthorized bool `json:"is_authorized" example:"true"`
}

// RefreshTokenRequest represents a request to refresh a token
type RefreshTokenRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称
	UserID     string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                        // 用户ID
	ServerName string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`                       // 服务名称
}

// RefreshTokenResponse represents a response for token refresh
type RefreshTokenResponse struct {
	Provider    string `json:"provider" example:"x"`
	UserID      string `json:"user_id" example:"user123"`
	ServerName  string `json:"server_name" example:"myapp"`
	ExpiresAt   int64  `json:"expires_at" example:"1704067199"`   // 新token的过期时间戳
	RefreshedAt int64  `json:"refreshed_at" example:"1704067199"` // 刷新时间戳
	Message     string `json:"message" example:"Token refreshed successfully"`
}

// CheckTokenStatusRequest represents a request to check token status
type CheckTokenStatusRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称
	UserID     string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                        // 用户ID
	ServerName string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`                       // 服务名称
}

// CheckTokenStatusResponse represents a response for token status check
type CheckTokenStatusResponse struct {
	Exists    bool   `json:"exists" example:"true"`            // token是否存在
	IsValid   bool   `json:"is_valid" example:"true"`          // token是否有效
	ExpiresAt int64  `json:"expires_at" example:"1704067199"`  // token过期时间戳
	Message   string `json:"message" example:"Token is valid"` // 状态消息
}

// GetRecentPostsRequest represents a request to get recent posts from a social platform
type GetRecentPostsRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称
	UserID     string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`                        // 用户ID
	ServerName string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"`                       // 服务名称
	Limit      int    `json:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`                    // 获取数量限制，默认10，最大100
	StartTime  int64  `json:"start_time,omitempty" example:"1704067199"`                                         // 开始时间戳（可选）
	EndTime    int64  `json:"end_time,omitempty" example:"1704153599"`                                           // 结束时间戳（可选）
}

// Post represents a single post from a social platform
type Post struct {
	ID          string    `json:"id" example:"1234567890"`                                      // 帖子ID
	Content     string    `json:"content" example:"Hello World!"`                               // 帖子内容
	MediaURL    string    `json:"media_url,omitempty" example:"https://example.com/image.jpg"`  // 媒体URL
	CreatedAt   int64     `json:"created_at" example:"1704067199"`                              // 创建时间戳
	UpdatedAt   int64     `json:"updated_at,omitempty" example:"1704067199"`                    // 更新时间戳
	Stats       StatsData `json:"stats"`                                                        // 统计信息
	URL         string    `json:"url,omitempty" example:"https://x.com/user/status/1234567890"` // 帖子链接
	MediaType   string    `json:"media_type,omitempty" example:"image"`                         // 媒体类型：image, video, audio
	Title       string    `json:"title,omitempty" example:"My Post"`                            // 标题（YouTube等平台）
	Description string    `json:"description,omitempty" example:"Post description"`             // 描述
	Tags        []string  `json:"tags" example:"tag1,tag2"`                                     // 标签列表
}

// GetRecentPostsResponse represents the response for recent posts
type GetRecentPostsResponse struct {
	Provider   string `json:"provider" example:"x"`
	UserID     string `json:"user_id" example:"user123"`
	ServerName string `json:"server_name" example:"myapp"`
	Posts      []Post `json:"posts"`              // 最近发布的帖子列表
	Total      int    `json:"total" example:"10"` // 总数量
}

// BatchGetRecentPostsRequest represents a request to get recent posts from multiple platforms
type BatchGetRecentPostsRequest struct {
	UserID     string `json:"user_id" binding:"required,min=1,max=100" example:"user123"`  // 用户ID
	ServerName string `json:"server_name" binding:"required,min=1,max=50" example:"myapp"` // 服务名称
	StartTime  int64  `json:"start_time,omitempty" example:"1704067199"`                   // 开始时间戳（可选）
	EndTime    int64  `json:"end_time,omitempty" example:"1704153599"`                     // 结束时间戳（可选）
	Platforms  []struct {
		Provider string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称
		Limit    int    `json:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`                    // 获取数量限制，默认10，最大100
	} `json:"platforms" binding:"required,min=1,max=10"` // 平台列表，最多10个平台
}

// PlatformPosts represents posts from a single platform
type PlatformPosts struct {
	Provider   string `json:"provider" example:"x"`
	UserID     string `json:"user_id" example:"user123"`
	ServerName string `json:"server_name" example:"myapp"`
	Posts      []Post `json:"posts"`                                           // 该平台的帖子列表
	Total      int    `json:"total" example:"10"`                              // 该平台的总数量
	Error      string `json:"error,omitempty" example:"authentication failed"` // 如果该平台查询失败，记录错误信息
}

// BatchGetRecentPostsResponse represents the response for batch recent posts
type BatchGetRecentPostsResponse struct {
	UserID       string          `json:"user_id" example:"user123"`
	ServerName   string          `json:"server_name" example:"myapp"`
	Platforms    []PlatformPosts `json:"platforms"`                 // 各平台的帖子列表
	TotalPosts   int             `json:"total_posts" example:"25"`  // 所有平台的总帖子数
	SuccessCount int             `json:"success_count" example:"3"` // 成功查询的平台数量
	ErrorCount   int             `json:"error_count" example:"1"`   // 查询失败的平台数量
}
