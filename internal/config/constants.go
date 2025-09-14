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
	TikTokAuthURL  = "https://www.tiktok.com/auth/authorize/"
	TikTokTokenURL = "https://open.tiktokapis.com/v2/oauth/token/"

	// Instagram OAuth endpoints (uses Facebook)
	InstagramAuthURL  = "https://www.facebook.com/v18.0/dialog/oauth"
	InstagramTokenURL = "https://graph.facebook.com/v18.0/oauth/access_token"
)

// Default OAuth scopes
var (
	DefaultYouTubeScopes = []string{
		"https://www.googleapis.com/auth/youtube.upload",
		"openid",
		"email",
	}

	DefaultXScopes = []string{
		"tweet.read",
		"tweet.write",
		"users.read",
		"offline.access",
	}

	DefaultFacebookScopes = []string{
		"pages_manage_posts",
		"pages_read_engagement",
		"pages_show_list",
		"pages_read_user_content",
	}

	DefaultTikTokScopes = []string{
		"video.upload",
		"user.info.basic",
	}

	DefaultInstagramScopes = []string{
		"instagram_content_publish",
		"pages_read_engagement",
	}
)

// Supported providers
var SupportedProviders = []string{
	"youtube",
	"x",
	"facebook",
	"tiktok",
	"instagram",
}

// Default configuration values
const (
	DefaultPort      = "8080"
	DefaultBaseURL   = "http://localhost:8080"
	DefaultRedisAddr = "localhost:6379"
	DefaultRedisDB   = 0
)
