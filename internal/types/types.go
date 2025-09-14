package types

import (
	"context"
	"net/http"
)

// ShareRequest represents a request to share content to a social platform
type ShareRequest struct {
	Provider   string   `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" validate:"required,oneof=youtube x facebook tiktok instagram" example:"x"`
	UserID     string   `json:"user_id" binding:"required,min=1,max=100" validate:"required,min=1,max=100" example:"user123"`
	ServerName string   `json:"server_name" binding:"required,min=1,max=50" validate:"required,min=1,max=50" example:"myapp"`
	Content    string   `json:"content,omitempty" binding:"max=280" validate:"max=280" example:"Hello World!"`                                // text content
	MediaURL   string   `json:"media_url,omitempty" binding:"omitempty,url" validate:"omitempty,url" example:"https://example.com/image.jpg"` // url to media (backend should download & upload)
	Title      string   `json:"title,omitempty" binding:"max=100" validate:"max=100" example:"My Post"`
	Desc       string   `json:"description,omitempty" binding:"max=500" validate:"max=500" example:"This is a description"`
	Tags       []string `json:"tags,omitempty" binding:"max=10" validate:"max=10" example:"hello,world"`
	Privacy    string   `json:"privacy,omitempty" binding:"omitempty,oneof=public private unlisted friends followers" validate:"omitempty,oneof=public private unlisted friends followers" example:"public"`
}

// StatsRequest represents a request to get statistics from a social platform
type StatsRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" validate:"required,oneof=youtube x facebook tiktok instagram" example:"x"`
	UserID     string `json:"user_id" binding:"required,min=1,max=100" validate:"required,min=1,max=100" example:"user123"`
	ServerName string `json:"server_name" binding:"required,min=1,max=50" validate:"required,min=1,max=50" example:"myapp"`
	MediaID    string `json:"media_id,omitempty" binding:"max=100" validate:"max=100" example:"1234567890"`
}

// StartAuthRequest represents a request to start OAuth authentication
type StartAuthRequest struct {
	Provider    string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" validate:"required,oneof=youtube x facebook tiktok instagram" example:"x"`
	UserID      string `json:"user_id" binding:"required,min=1,max=100" validate:"required,min=1,max=100" example:"user123"`
	RedirectURI string `json:"redirect_uri" binding:"required,url" validate:"required,url" example:"https://myapp.com/auth/callback"`
	ServerName  string `json:"server_name" binding:"required,min=1,max=50" validate:"required,min=1,max=50" example:"myapp"`
}

// CallbackRequest represents a request for OAuth callback
// 前端收到OAuth回调后，调用此接口处理授权码交换
type CallbackRequest struct {
	Provider    string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" validate:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称
	ServerName  string `json:"server_name" binding:"required,min=1,max=50" validate:"required,min=1,max=50" example:"myapp"`                                                    // 服务器名称
	State       string `json:"state" binding:"required,min=1" validate:"required,min=1" example:"encoded_state_string"`                                                         // 状态参数，包含用户ID等信息
	Code        string `json:"code" binding:"required,min=1" validate:"required,min=1" example:"authorization_code"`                                                            // 授权码
	RedirectURI string `json:"redirect_uri" binding:"required,url" validate:"required,url" example:"https://myapp.com/auth/callback"`                                           // 重定向URI
}

// RefreshTokenRequest represents a request for refreshing access token
type RefreshTokenRequest struct {
	Provider   string `json:"provider" binding:"required,oneof=youtube x facebook tiktok instagram" validate:"required,oneof=youtube x facebook tiktok instagram" example:"x"` // 平台名称
	UserID     string `json:"user_id" binding:"required,min=1,max=100" validate:"required,min=1,max=100" example:"user123"`                                                    // 用户ID
	ServerName string `json:"server_name" binding:"required,min=1,max=50" validate:"required,min=1,max=50" example:"myapp"`                                                    // 服务器名称
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
	Provider   string    `json:"provider" example:"x"`
	UserID     string    `json:"user_id" example:"user123"`
	ServerName string    `json:"server_name" example:"myapp"`
	ExpiresAt  int64     `json:"expires_at" example:"1704067199"` // 时间戳格式
	ReferAt    int64     `json:"refer_at" example:"1704067199"`   // 时间戳格式
	Message    string    `json:"message" example:"OAuth callback completed successfully"`
}

// RefreshTokenResponse represents the response for token refresh
type RefreshTokenResponse struct {
	Provider   string `json:"provider" example:"x"`
	UserID     string `json:"user_id" example:"user123"`
	ServerName string `json:"server_name" example:"myapp"`
	ExpiresAt  int64  `json:"expires_at" example:"1704067199"` // 时间戳格式
	ReferAt    int64  `json:"refer_at" example:"1704067199"`   // 时间戳格式
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

// StatsResponse represents the response for statistics
type StatsResponse struct {
	Provider   string                 `json:"provider" example:"x"`
	UserID     string                 `json:"user_id" example:"user123"`
	ServerName string                 `json:"server_name" example:"myapp"`
	MediaID    string                 `json:"media_id" example:"1234567890"`
	Stats      map[string]interface{} `json:"stats" example:"{\"likes\": 100, \"retweets\": 50, \"replies\": 25}"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// Platform represents a social media platform interface
type Platform interface {
	// Share shares content to the platform
	Share(ctx context.Context, client *http.Client, req *ShareRequest) error

	// GetStats retrieves statistics from the platform
	GetStats(ctx context.Context, client *http.Client, mediaID string) (map[string]interface{}, error)

	// GetName returns the platform name
	GetName() string

	// HandleOAuthCallback handles OAuth callback for the platform
	HandleOAuthCallback(ctx context.Context, code, state string) error
}
