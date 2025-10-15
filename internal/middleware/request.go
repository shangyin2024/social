package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	ctxutil "social/pkg/context"
	"social/pkg/logger"
)

// RequestMiddleware handles request ID generation and logging
type RequestMiddleware struct {
	logger *logger.Logger
}

// NewRequestMiddleware creates a new request middleware
func NewRequestMiddleware(logger *logger.Logger) *RequestMiddleware {
	return &RequestMiddleware{
		logger: logger,
	}
}

// RequestID creates a middleware that generates request ID and adds it to context
func (m *RequestMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()

		// Add request ID to context
		ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add request ID to response header
		c.Header("X-Request-ID", requestID)

		// Log request start
		m.logger.Info(ctx, "request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request completion
		duration := time.Since(start)
		m.logger.Info(ctx, "request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	}
}
