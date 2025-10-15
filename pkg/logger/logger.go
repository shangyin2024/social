package logger

import (
	"context"
	"log/slog"
	"os"

	ctxutil "social/pkg/context"
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

// Error logs an error with context, automatically extracting request ID
func (l *Logger) Error(ctx context.Context, err error, message string, args ...interface{}) {
	// Extract request ID from context if available
	if requestID, ok := ctxutil.GetRequestID(ctx); ok {
		args = append([]interface{}{"request_id", requestID}, args...)
	}
	l.ErrorContext(ctx, message, append([]interface{}{"error", err}, args...)...)
}

// Info logs an info message with context, automatically extracting request ID
func (l *Logger) Info(ctx context.Context, message string, args ...interface{}) {
	// Extract request ID from context if available
	if requestID, ok := ctxutil.GetRequestID(ctx); ok {
		args = append([]interface{}{"request_id", requestID}, args...)
	}
	l.InfoContext(ctx, message, args...)
}

// Warn logs a warning message with context, automatically extracting request ID
func (l *Logger) Warn(ctx context.Context, message string, args ...interface{}) {
	// Extract request ID from context if available
	if requestID, ok := ctxutil.GetRequestID(ctx); ok {
		args = append([]interface{}{"request_id", requestID}, args...)
	}
	l.WarnContext(ctx, message, args...)
}
