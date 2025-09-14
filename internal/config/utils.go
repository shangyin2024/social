package config

import (
	"fmt"
	"net/url"
	"strings"
)

// BuildRedirectURI constructs a redirect URI for OAuth callback
func (c *Config) BuildRedirectURI(customRedirectURI string) string {
	// If custom redirect URI is provided and it's a full URL, use it
	if customRedirectURI != "" && strings.HasPrefix(customRedirectURI, "http") {
		return customRedirectURI
	}

	// If custom redirect URI is provided but not a full URL, treat it as a path
	if customRedirectURI != "" {
		return fmt.Sprintf("%s%s", c.Server.BaseURL, customRedirectURI)
	}

	// Default callback path
	return fmt.Sprintf("%s/static/callback.html", c.Server.BaseURL)
}

// ValidateRedirectURI validates if a redirect URI is safe and properly formatted
func (c *Config) ValidateRedirectURI(redirectURI string) error {
	if redirectURI == "" {
		return fmt.Errorf("redirect URI cannot be empty")
	}

	// Parse the URL
	u, err := url.Parse(redirectURI)
	if err != nil {
		return fmt.Errorf("invalid redirect URI format: %w", err)
	}

	// Check if it's a valid scheme
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("redirect URI must use http or https scheme")
	}

	// For security, ensure the redirect URI is from the same domain or localhost
	baseURL, err := url.Parse(c.Server.BaseURL)
	if err != nil {
		return fmt.Errorf("invalid base URL in config: %w", err)
	}

	// Allow localhost for development
	if u.Host == "localhost" || u.Host == "127.0.0.1" {
		return nil
	}

	// Allow same domain
	if u.Host == baseURL.Host {
		return nil
	}

	// Allow subdomains of the base domain
	if strings.HasSuffix(u.Host, "."+baseURL.Host) {
		return nil
	}

	return fmt.Errorf("redirect URI host '%s' is not allowed, must be same domain as base URL '%s' or localhost", u.Host, baseURL.Host)
}

// GetCallbackURL returns the callback URL for a specific provider
func (c *Config) GetCallbackURL(provider string) string {
	return fmt.Sprintf("%s/auth/%s/callback", c.Server.BaseURL, provider)
}

// GetStaticCallbackURL returns the static callback page URL
func (c *Config) GetStaticCallbackURL() string {
	return fmt.Sprintf("%s/static/callback.html", c.Server.BaseURL)
}

// IsDevelopmentMode checks if the application is running in development mode
func (c *Config) IsDevelopmentMode() bool {
	return strings.Contains(c.Server.BaseURL, "localhost") ||
		strings.Contains(c.Server.BaseURL, "127.0.0.1") ||
		strings.Contains(c.Server.BaseURL, "dev") ||
		strings.Contains(c.Server.BaseURL, "test")
}

// GetEnvironment returns the environment type based on base URL
func (c *Config) GetEnvironment() string {
	if c.IsDevelopmentMode() {
		return "development"
	}
	if strings.Contains(c.Server.BaseURL, "staging") {
		return "staging"
	}
	return "production"
}
