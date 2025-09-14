package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	googleoauth "golang.org/x/oauth2/google"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig                 `mapstructure:"server"`
	Redis    RedisConfig                  `mapstructure:"redis"`
	OAuth    OAuthConfig                  `mapstructure:"oauth"`
	Platform PlatformConfig               `mapstructure:"platform"`
	Servers  map[string]ServerOAuthConfig `mapstructure:"servers"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port    string `mapstructure:"port"`
	BaseURL string `mapstructure:"base_url"`
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// OAuthConfig holds OAuth provider configurations
type OAuthConfig struct {
	YouTube   ProviderConfig `mapstructure:"youtube"`
	X         ProviderConfig `mapstructure:"x"`
	Facebook  ProviderConfig `mapstructure:"facebook"`
	TikTok    ProviderConfig `mapstructure:"tiktok"`
	Instagram ProviderConfig `mapstructure:"instagram"`
}

// ProviderConfig holds configuration for a single OAuth provider
type ProviderConfig struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	Scopes       []string `mapstructure:"scopes"`
}

// PlatformConfig holds platform-specific settings
type PlatformConfig struct {
	SupportedProviders []string `mapstructure:"supported_providers"`
}

// ServerOAuthConfig holds OAuth configuration for a specific server
type ServerOAuthConfig struct {
	YouTube   ProviderConfig `mapstructure:"youtube"`
	X         ProviderConfig `mapstructure:"x"`
	Facebook  ProviderConfig `mapstructure:"facebook"`
	TikTok    ProviderConfig `mapstructure:"tiktok"`
	Instagram ProviderConfig `mapstructure:"instagram"`
}

// Load loads configuration from environment variables and files
func Load() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults()

	// Try to load environment-specific config file first
	configFile := GetConfigFile()
	viper.SetConfigName(strings.TrimSuffix(configFile, ".yaml"))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/social")

	// Try to read environment-specific config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}

		// If environment-specific config not found, try default config
		if configFile != "config.yaml" {
			viper.SetConfigName("config")
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					return nil, fmt.Errorf("failed to read default config file: %w", err)
				}
				// No config file found, continue with environment variables only
			}
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables if set
	overrideWithEnvVars(&config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// overrideWithEnvVars overrides config values with environment variables
func overrideWithEnvVars(config *Config) {
	if port := GetEnvWithDefault(EnvServerPort, ""); port != "" {
		config.Server.Port = port
	}
	if baseURL := GetEnvWithDefault(EnvServerBaseURL, ""); baseURL != "" {
		config.Server.BaseURL = baseURL
	}
	if redisAddr := GetEnvWithDefault(EnvRedisAddr, ""); redisAddr != "" {
		config.Redis.Addr = redisAddr
	}
	if redisPassword := GetEnvWithDefault(EnvRedisPassword, ""); redisPassword != "" {
		config.Redis.Password = redisPassword
	}
	if redisDB := GetEnvWithDefault(EnvRedisDB, ""); redisDB != "" {
		// Note: viper will handle the string to int conversion
		config.Redis.DB = 0 // This will be overridden by viper if env var is set
	}
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("server.port", DefaultPort)
	viper.SetDefault("server.base_url", DefaultBaseURL)
	viper.SetDefault("redis.addr", DefaultRedisAddr)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", DefaultRedisDB)
	viper.SetDefault("platform.supported_providers", SupportedProviders)

	// OAuth scopes defaults
	viper.SetDefault("oauth.youtube.scopes", DefaultYouTubeScopes)
	viper.SetDefault("oauth.x.scopes", DefaultXScopes)
	viper.SetDefault("oauth.facebook.scopes", DefaultFacebookScopes)
	viper.SetDefault("oauth.tiktok.scopes", DefaultTikTokScopes)
	viper.SetDefault("oauth.instagram.scopes", DefaultInstagramScopes)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	validator := NewConfigValidator(c)
	return validator.ValidateAll()
}

// GetWarnings returns configuration warnings
func (c *Config) GetWarnings() []string {
	validator := NewConfigValidator(c)
	return validator.GetValidationWarnings()
}

// GetOAuthConfig returns oauth2.Config for the specified provider
func (c *Config) GetOAuthConfig(provider string) (*oauth2.Config, error) {
	baseURL := c.Server.BaseURL

	switch provider {
	case "youtube":
		return &oauth2.Config{
			ClientID:     c.OAuth.YouTube.ClientID,
			ClientSecret: c.OAuth.YouTube.ClientSecret,
			Scopes:       c.OAuth.YouTube.Scopes,
			Endpoint:     googleoauth.Endpoint,
			RedirectURL:  fmt.Sprintf("%s/auth/youtube/callback", baseURL),
		}, nil
	case "x":
		return &oauth2.Config{
			ClientID:     c.OAuth.X.ClientID,
			ClientSecret: c.OAuth.X.ClientSecret,
			Scopes:       c.OAuth.X.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  XAuthURL,
				TokenURL: XTokenURL,
			},
			RedirectURL: fmt.Sprintf("%s/auth/x/callback", baseURL),
		}, nil
	case "facebook":
		return &oauth2.Config{
			ClientID:     c.OAuth.Facebook.ClientID,
			ClientSecret: c.OAuth.Facebook.ClientSecret,
			Scopes:       c.OAuth.Facebook.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  FacebookAuthURL,
				TokenURL: FacebookTokenURL,
			},
			RedirectURL: fmt.Sprintf("%s/auth/facebook/callback", baseURL),
		}, nil
	case "tiktok":
		return &oauth2.Config{
			ClientID:     c.OAuth.TikTok.ClientID,
			ClientSecret: c.OAuth.TikTok.ClientSecret,
			Scopes:       c.OAuth.TikTok.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  TikTokAuthURL,
				TokenURL: TikTokTokenURL,
			},
			RedirectURL: fmt.Sprintf("%s/auth/tiktok/callback", baseURL),
		}, nil
	case "instagram":
		return &oauth2.Config{
			ClientID:     c.OAuth.Instagram.ClientID,
			ClientSecret: c.OAuth.Instagram.ClientSecret,
			Scopes:       c.OAuth.Instagram.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  InstagramAuthURL,
				TokenURL: InstagramTokenURL,
			},
			RedirectURL: fmt.Sprintf("%s/auth/instagram/callback", baseURL),
		}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// GetOAuthConfigWithRedirect returns oauth2.Config for the specified provider with custom redirect URI
func (c *Config) GetOAuthConfigWithRedirect(provider, redirectURI string) (*oauth2.Config, error) {
	switch provider {
	case "youtube":
		return &oauth2.Config{
			ClientID:     c.OAuth.YouTube.ClientID,
			ClientSecret: c.OAuth.YouTube.ClientSecret,
			Scopes:       c.OAuth.YouTube.Scopes,
			Endpoint:     googleoauth.Endpoint,
			RedirectURL:  redirectURI,
		}, nil
	case "x":
		return &oauth2.Config{
			ClientID:     c.OAuth.X.ClientID,
			ClientSecret: c.OAuth.X.ClientSecret,
			Scopes:       c.OAuth.X.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  XAuthURL,
				TokenURL: XTokenURL,
			},
			RedirectURL: redirectURI,
		}, nil
	case "facebook":
		return &oauth2.Config{
			ClientID:     c.OAuth.Facebook.ClientID,
			ClientSecret: c.OAuth.Facebook.ClientSecret,
			Scopes:       c.OAuth.Facebook.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  FacebookAuthURL,
				TokenURL: FacebookTokenURL,
			},
			RedirectURL: redirectURI,
		}, nil
	case "tiktok":
		return &oauth2.Config{
			ClientID:     c.OAuth.TikTok.ClientID,
			ClientSecret: c.OAuth.TikTok.ClientSecret,
			Scopes:       c.OAuth.TikTok.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  TikTokAuthURL,
				TokenURL: TikTokTokenURL,
			},
			RedirectURL: redirectURI,
		}, nil
	case "instagram":
		return &oauth2.Config{
			ClientID:     c.OAuth.Instagram.ClientID,
			ClientSecret: c.OAuth.Instagram.ClientSecret,
			Scopes:       c.OAuth.Instagram.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  InstagramAuthURL,
				TokenURL: InstagramTokenURL,
			},
			RedirectURL: redirectURI,
		}, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

// GetServerOAuthConfig returns oauth2.Config for the specified provider and server
func (c *Config) GetServerOAuthConfig(provider, serverName, redirectURI string) (*oauth2.Config, error) {
	// 首先尝试从服务器特定配置获取
	if serverConfig, exists := c.Servers[serverName]; exists {
		switch provider {
		case "youtube":
			return &oauth2.Config{
				ClientID:     serverConfig.YouTube.ClientID,
				ClientSecret: serverConfig.YouTube.ClientSecret,
				Scopes:       serverConfig.YouTube.Scopes,
				Endpoint:     googleoauth.Endpoint,
				RedirectURL:  redirectURI,
			}, nil
		case "x":
			return &oauth2.Config{
				ClientID:     serverConfig.X.ClientID,
				ClientSecret: serverConfig.X.ClientSecret,
				Scopes:       serverConfig.X.Scopes,
				Endpoint: oauth2.Endpoint{
					AuthURL:  XAuthURL,
					TokenURL: XTokenURL,
				},
				RedirectURL: redirectURI,
			}, nil
		case "facebook":
			return &oauth2.Config{
				ClientID:     serverConfig.Facebook.ClientID,
				ClientSecret: serverConfig.Facebook.ClientSecret,
				Scopes:       serverConfig.Facebook.Scopes,
				Endpoint: oauth2.Endpoint{
					AuthURL:  FacebookAuthURL,
					TokenURL: FacebookTokenURL,
				},
				RedirectURL: redirectURI,
			}, nil
		case "tiktok":
			return &oauth2.Config{
				ClientID:     serverConfig.TikTok.ClientID,
				ClientSecret: serverConfig.TikTok.ClientSecret,
				Scopes:       serverConfig.TikTok.Scopes,
				Endpoint: oauth2.Endpoint{
					AuthURL:  TikTokAuthURL,
					TokenURL: TikTokTokenURL,
				},
				RedirectURL: redirectURI,
			}, nil
		case "instagram":
			return &oauth2.Config{
				ClientID:     serverConfig.Instagram.ClientID,
				ClientSecret: serverConfig.Instagram.ClientSecret,
				Scopes:       serverConfig.Instagram.Scopes,
				Endpoint: oauth2.Endpoint{
					AuthURL:  InstagramAuthURL,
					TokenURL: InstagramTokenURL,
				},
				RedirectURL: redirectURI,
			}, nil
		}
	}

	// 如果服务器特定配置不存在，回退到默认配置，使用动态redirect_uri
	return c.GetOAuthConfigWithRedirect(provider, redirectURI)
}

// IsProviderSupported checks if a provider is supported
func (c *Config) IsProviderSupported(provider string) bool {
	for _, supported := range c.Platform.SupportedProviders {
		if supported == provider {
			return true
		}
	}
	return false
}
