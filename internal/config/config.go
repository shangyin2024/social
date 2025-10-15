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
	Server  ServerConfig                 `mapstructure:"server"`
	Redis   RedisConfig                  `mapstructure:"redis"`
	Servers map[string]ServerOAuthConfig `mapstructure:"servers"`
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

// ProviderConfig holds configuration for a single OAuth provider
type ProviderConfig struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	Scopes       []string `mapstructure:"scopes"`
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
}

// Validate validates the configuration
func (c *Config) Validate() error {
	validator := NewConfigValidator(c)
	return validator.ValidateAll()
}

// GetServerOAuthConfig returns oauth2.Config for the specified provider and server
func (c *Config) GetServerOAuthConfig(provider, serverName, redirectURI string) (*oauth2.Config, error) {
	// 从服务器特定配置获取
	serverConfig, exists := c.Servers[serverName]
	if !exists {
		return nil, fmt.Errorf("server configuration not found: %s", serverName)
	}

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
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
