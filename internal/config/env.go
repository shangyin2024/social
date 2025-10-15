package config

import (
	"os"
	"strings"
)

// Environment variables
const (
	EnvServerPort    = "SERVER_PORT"
	EnvServerBaseURL = "SERVER_BASE_URL"
	EnvRedisAddr     = "REDIS_ADDR"
	EnvRedisPassword = "REDIS_PASSWORD"
	EnvRedisDB       = "REDIS_DB"
	EnvGinMode       = "GIN_MODE"
)

// GetEnvWithDefault returns environment variable value or default if not set
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvBool returns boolean environment variable value or default if not set
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// GetEnvironment returns the current environment type
func GetEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = os.Getenv("APP_ENV")
	}

	switch strings.ToLower(env) {
	case "development", "dev":
		return "development"
	case "staging", "stage":
		return "staging"
	case "production", "prod":
		return "production"
	default:
		return "development"
	}
}

// IsDevelopment returns true if running in development mode
func IsDevelopment() bool {
	return GetEnvironment() == "development"
}

// IsProduction returns true if running in production mode
func IsProduction() bool {
	return GetEnvironment() == "production"
}

// IsStaging returns true if running in staging mode
func IsStaging() bool {
	return GetEnvironment() == "staging"
}

// GetConfigFile returns the appropriate config file name based on environment
func GetConfigFile() string {
	env := GetEnvironment()
	switch env {
	case "development":
		return "config.dev.yaml"
	case "staging":
		return "config.staging.yaml"
	case "production":
		return "config.prod.yaml"
	default:
		return "config.yaml"
	}
}

// GetLogLevel returns the appropriate log level based on environment
func GetLogLevel() string {
	if IsDevelopment() {
		return "debug"
	}
	return "info"
}
