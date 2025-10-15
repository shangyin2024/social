package config

// OAuth provider endpoints
const (
	// YouTube OAuth endpoints
	YouTubeAuthURL  = "https://accounts.google.com/o/oauth2/auth"
	YouTubeTokenURL = "https://oauth2.googleapis.com/token"

	// X (Twitter) OAuth endpoints
	XAuthURL  = "https://x.com/i/oauth2/authorize"
	XTokenURL = "https://api.x.com/2/oauth2/token"

	// Facebook OAuth endpoints
	FacebookAuthURL  = "https://www.facebook.com/v18.0/dialog/oauth"
	FacebookTokenURL = "https://graph.facebook.com/v18.0/oauth/access_token"

	// TikTok OAuth endpoints
	TikTokAuthURL  = "https://www.tiktok.com/v2/auth/authorize/"
	TikTokTokenURL = "https://open.tiktokapis.com/v2/oauth/token/"

	// Instagram OAuth endpoints
	InstagramAuthURL  = "https://api.instagram.com/oauth/authorize"
	InstagramTokenURL = "https://api.instagram.com/oauth/access_token"
)

// Default configuration values
const (
	DefaultPort      = "8080"
	DefaultBaseURL   = "http://localhost:8080"
	DefaultRedisAddr = "localhost:6379"
	DefaultRedisDB   = 0
)
