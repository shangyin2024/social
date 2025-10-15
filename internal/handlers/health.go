package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"social/internal/storage"
	ctxutil "social/pkg/context"
	"social/pkg/logger"
	"social/pkg/response"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	storage storage.Storage
	logger  *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(storage storage.Storage, logger *logger.Logger) *HealthHandler {
	return &HealthHandler{
		storage: storage,
		logger:  logger,
	}
}

// Health performs a health check
// @Summary 健康检查
// @Description 检查服务健康状态，包括存储连接状态
// @Tags 系统
// @Produce json
// @Success 200 {object} map[string]any "健康状态"
// @Failure 503 {object} types.ErrorResponse "服务不可用"
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	// Check Redis connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.storage.Health(ctx); err != nil {
		h.logger.Error(ctx, err, "health check failed - storage unavailable")
		response.ServiceUnavailable(c, "service unavailable")
		return
	}

	response.Success(c, gin.H{
		"timestamp": time.Now().UTC(),
	})
}
