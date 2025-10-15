package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ConfigValidator provides configuration validation functionality
type ConfigValidator struct {
	config *Config
}

// NewConfigValidator creates a new config validator
func NewConfigValidator(config *Config) *ConfigValidator {
	return &ConfigValidator{config: config}
}

// ValidateAll performs comprehensive configuration validation
func (v *ConfigValidator) ValidateAll() error {
	if err := v.ValidateServer(); err != nil {
		return fmt.Errorf("server validation failed: %w", err)
	}

	if err := v.ValidateRedis(); err != nil {
		return fmt.Errorf("redis validation failed: %w", err)
	}

	if err := v.ValidateOAuth(); err != nil {
		return fmt.Errorf("oauth validation failed: %w", err)
	}

	if err := v.ValidateServers(); err != nil {
		return fmt.Errorf("servers validation failed: %w", err)
	}

	return nil
}

// ValidateServer validates server configuration
func (v *ConfigValidator) ValidateServer() error {
	if v.config.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if v.config.Server.BaseURL == "" {
		return fmt.Errorf("server base URL is required")
	}

	// Validate base URL format
	if _, err := url.Parse(v.config.Server.BaseURL); err != nil {
		return fmt.Errorf("invalid base URL format: %w", err)
	}

	// Validate port format
	portRegex := regexp.MustCompile(`^\d+$`)
	if !portRegex.MatchString(v.config.Server.Port) {
		return fmt.Errorf("invalid port format: %s", v.config.Server.Port)
	}

	return nil
}

// ValidateRedis validates Redis configuration
func (v *ConfigValidator) ValidateRedis() error {
	if v.config.Redis.Addr == "" {
		return fmt.Errorf("redis address is required")
	}

	// Validate Redis address format (host:port)
	parts := strings.Split(v.config.Redis.Addr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid redis address format: %s", v.config.Redis.Addr)
	}

	// Validate port
	portRegex := regexp.MustCompile(`^\d+$`)
	if !portRegex.MatchString(parts[1]) {
		return fmt.Errorf("invalid redis port format: %s", parts[1])
	}

	return nil
}

// ValidateOAuth validates OAuth configuration in servers
func (v *ConfigValidator) ValidateOAuth() error {
	// 验证每个服务器的 OAuth 配置
	for serverName, serverConfig := range v.config.Servers {
		providers := map[string]ProviderConfig{
			"youtube":   serverConfig.YouTube,
			"x":         serverConfig.X,
			"facebook":  serverConfig.Facebook,
			"tiktok":    serverConfig.TikTok,
			"instagram": serverConfig.Instagram,
		}

		for name, provider := range providers {
			if err := v.ValidateProvider(name, provider); err != nil {
				return fmt.Errorf("server %s: %w", serverName, err)
			}
		}
	}

	return nil
}

// ValidateProvider validates a single OAuth provider configuration
func (v *ConfigValidator) ValidateProvider(name string, provider ProviderConfig) error {
	if provider.ClientID == "" {
		return fmt.Errorf("OAuth provider %s client ID is required", name)
	}

	if provider.ClientSecret == "" {
		return fmt.Errorf("OAuth provider %s client secret is required", name)
	}

	if len(provider.Scopes) == 0 {
		return fmt.Errorf("OAuth provider %s scopes are required", name)
	}

	// Validate client ID format (basic check)
	if len(provider.ClientID) < 10 {
		return fmt.Errorf("OAuth provider %s client ID seems too short", name)
	}

	// Validate client secret format (basic check)
	if len(provider.ClientSecret) < 10 {
		return fmt.Errorf("OAuth provider %s client secret seems too short", name)
	}

	return nil
}

// ValidateServers validates multi-server configuration
func (v *ConfigValidator) ValidateServers() error {
	for serverName, serverConfig := range v.config.Servers {
		if err := v.ValidateServerConfig(serverName, serverConfig); err != nil {
			return err
		}
	}
	return nil
}

// ValidateServerConfig validates a single server configuration
func (v *ConfigValidator) ValidateServerConfig(serverName string, serverConfig ServerOAuthConfig) error {
	if serverName == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Validate server name format (alphanumeric and hyphens only)
	serverNameRegex := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !serverNameRegex.MatchString(serverName) {
		return fmt.Errorf("invalid server name format: %s", serverName)
	}

	// Validate each provider in server config
	providers := map[string]ProviderConfig{
		"youtube":   serverConfig.YouTube,
		"x":         serverConfig.X,
		"facebook":  serverConfig.Facebook,
		"tiktok":    serverConfig.TikTok,
		"instagram": serverConfig.Instagram,
	}

	for providerName, provider := range providers {
		// Only validate if provider is configured (not empty)
		if provider.ClientID != "" || provider.ClientSecret != "" {
			if err := v.ValidateProvider(fmt.Sprintf("%s.%s", serverName, providerName), provider); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetValidationWarnings returns non-critical validation warnings
func (v *ConfigValidator) GetValidationWarnings() []string {
	var warnings []string

	// Check for development settings in production
	if IsProduction() {
		if strings.Contains(v.config.Server.BaseURL, "localhost") {
			warnings = append(warnings, "Using localhost in production environment")
		}
		if v.config.Server.Port == "8080" {
			warnings = append(warnings, "Using default port 8080 in production")
		}
	}

	// Check for missing server configurations
	if len(v.config.Servers) == 0 {
		warnings = append(warnings, "No multi-server configurations found")
	}

	// Check for empty OAuth configurations in servers
	for serverName, serverConfig := range v.config.Servers {
		providers := map[string]ProviderConfig{
			"youtube":   serverConfig.YouTube,
			"x":         serverConfig.X,
			"facebook":  serverConfig.Facebook,
			"tiktok":    serverConfig.TikTok,
			"instagram": serverConfig.Instagram,
		}

		for name, provider := range providers {
			if provider.ClientID == "" || provider.ClientSecret == "" {
				warnings = append(warnings, fmt.Sprintf("Server %s: OAuth provider %s is not configured", serverName, name))
			}
		}
	}

	return warnings
}
