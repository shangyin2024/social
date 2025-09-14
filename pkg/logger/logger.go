package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{Logger: logger}
}

// WithRequestID adds request ID to the logger context
func (l *Logger) WithRequestID(ctx context.Context, requestID string) *slog.Logger {
	return l.With("request_id", requestID)
}

// WithUserID adds user ID to the logger context
func (l *Logger) WithUserID(userID string) *slog.Logger {
	return l.With("user_id", userID)
}

// WithProvider adds provider to the logger context
func (l *Logger) WithProvider(provider string) *slog.Logger {
	return l.With("provider", provider)
}

// LogError logs an error with context
func (l *Logger) LogError(ctx context.Context, err error, message string, args ...interface{}) {
	l.ErrorContext(ctx, message, append([]interface{}{"error", err}, args...)...)
}

// LogInfo logs an info message with context
func (l *Logger) LogInfo(ctx context.Context, message string, args ...interface{}) {
	l.InfoContext(ctx, message, args...)
}

// LogWarn logs a warning message with context
func (l *Logger) LogWarn(ctx context.Context, message string, args ...interface{}) {
	l.WarnContext(ctx, message, args...)
}
