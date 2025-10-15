package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"social/internal/config"
	"social/internal/oauth"
	"social/internal/platforms"
	"social/internal/storage"
	"social/internal/types"
	"social/pkg/errors"
	"social/pkg/logger"
	"social/pkg/response"
)

// ShareHandler handles content sharing requests
type ShareHandler struct {
	config       *config.Config
	storage      storage.Storage
	registry     *platforms.Registry
	logger       *logger.Logger
	tokenManager *oauth.TokenManager
}

// NewShareHandler creates a new share handler
func NewShareHandler(cfg *config.Config, storage storage.Storage, registry *platforms.Registry, logger *logger.Logger) *ShareHandler {
	return &ShareHandler{
		config:       cfg,
		storage:      storage,
		registry:     registry,
		logger:       logger,
		tokenManager: oauth.NewTokenManager(cfg, storage, logger),
	}
}

// Share handles share requests
// @Summary 分享内容到社交媒体平台
// @Description 将内容分享到指定的社交媒体平台
// @Tags 分享
// @Accept json
// @Produce json
// @Param request body types.ShareRequest true "分享请求参数"
// @Success 200 {object} types.APIResponse{data=types.ShareResponse} "分享成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/share [post]
func (h *ShareHandler) Share(c *gin.Context) {
	ctx := c.Request.Context()

	var req types.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind share request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Get authenticated client with automatic token refresh
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client, err := h.tokenManager.CreateAuthenticatedClient(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to create authenticated client", "provider", req.Provider, "user_id", req.UserID)
		if err.Error() == "token not found" {
			response.Error(c, errors.ErrTokenNotFound)
		} else {
			response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("authentication failed: %v", err))
		}
		return
	}

	// Get platform implementation
	platform, err := h.registry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.Error(ctx, err, "platform not found", "provider", req.Provider)
		response.Error(c, errors.ErrPlatformNotSupported)
		return
	}

	// Check account status before sharing (for X platform)
	if req.Provider == "x" {
		h.logger.Info(ctx, "checking account status", "provider", req.Provider, "user_id", req.UserID)
		if xPlatform, ok := platform.(*platforms.XPlatform); ok {
			if err := xPlatform.CheckAccountStatus(ctx, client); err != nil {
				h.logger.Error(ctx, err, "account status check failed", "provider", req.Provider, "user_id", req.UserID)
				// Return a more specific error for account issues
				if strings.Contains(err.Error(), "suspended") {
					response.ErrorWithDetail(c, errors.ErrInternalServer, "账户已被暂停，请联系 X (Twitter) 客服解决")
				} else {
					response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("账户状态检查失败: %v", err))
				}
				return
			}
		}
	}

	// Share content
	h.logger.Info(ctx, "sharing content", "provider", req.Provider, "user_id", req.UserID)
	mediaID, err := platform.Share(ctx, client, &req)
	if err != nil {
		h.logger.Error(ctx, err, "failed to share content", "provider", req.Provider, "user_id", req.UserID)

		// Provide more specific error messages based on error type
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "account suspended") {
			response.ErrorWithDetail(c, errors.ErrInternalServer, "账户已被暂停，请联系 X (Twitter) 客服解决")
		} else if strings.Contains(errorMsg, "authentication failed") {
			response.ErrorWithDetail(c, errors.ErrInternalServer, "认证失败，请重新授权")
		} else if strings.Contains(errorMsg, "rate limit exceeded") {
			response.ErrorWithDetail(c, errors.ErrInternalServer, "请求过于频繁，请稍后再试")
		} else {
			response.ErrorWithDetail(c, errors.ErrInternalServer, errorMsg)
		}
		return
	}

	h.logger.Info(ctx, "content shared successfully", "provider", req.Provider, "user_id", req.UserID)

	shareResponse := types.ShareResponse{
		Provider:   req.Provider,
		UserID:     req.UserID,
		ServerName: req.ServerName,
		Content:    req.Content,
		MediaURL:   req.MediaURL,
		Tags:       req.Tags,
		MediaID:    mediaID,
	}
	response.SuccessWithMessage(c, "content shared successfully", shareResponse)
}

// GetStats handles statistics requests
// @Summary 获取社交媒体内容统计信息
// @Description 获取指定媒体内容在社交媒体平台上的统计信息
// @Tags 统计
// @Accept json
// @Produce json
// @Param request body types.StatsRequest true "统计请求参数"
// @Success 200 {object} types.APIResponse{data=types.StatsResponse} "统计信息"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/stats [post]
func (h *ShareHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	var req types.StatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind stats request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Get authenticated client with automatic token refresh
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := h.tokenManager.CreateAuthenticatedClient(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to create authenticated client", "provider", req.Provider, "user_id", req.UserID)
		if err.Error() == "token not found" {
			response.Error(c, errors.ErrTokenNotFound)
		} else {
			response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("authentication failed: %v", err))
		}
		return
	}

	// Get platform implementation
	platform, err := h.registry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.Error(ctx, err, "platform not found", "provider", req.Provider)
		response.Error(c, errors.ErrPlatformNotSupported)
		return
	}

	// Get statistics
	h.logger.Info(ctx, "getting statistics", "provider", req.Provider, "user_id", req.UserID, "media_id", req.MediaID)
	stats, err := platform.GetStats(ctx, client, req.MediaID)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get statistics", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, err.Error())
		return
	}

	h.logger.Info(ctx, "statistics retrieved successfully", "provider", req.Provider, "user_id", req.UserID)

	statsResponse := types.StatsResponse{
		Provider:   req.Provider,
		UserID:     req.UserID,
		ServerName: req.ServerName,
		MediaID:    req.MediaID,
		Stats:      stats,
	}
	response.Success(c, statsResponse)
}

// GetRecentPosts handles recent posts requests
// @Summary 获取最近发布的内容
// @Description 获取指定平台最近发布的内容列表
// @Tags 内容
// @Accept json
// @Produce json
// @Param request body types.GetRecentPostsRequest true "获取最近发布内容请求参数"
// @Success 200 {object} types.APIResponse{data=types.GetRecentPostsResponse} "获取成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/recent-posts [post]
func (h *ShareHandler) GetRecentPosts(c *gin.Context) {
	ctx := c.Request.Context()

	var req types.GetRecentPostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind recent posts request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Set default limit if not provided
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Get authenticated client with automatic token refresh
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client, err := h.tokenManager.CreateAuthenticatedClient(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to create authenticated client", "provider", req.Provider, "user_id", req.UserID)
		if err.Error() == "token not found" {
			response.Error(c, errors.ErrTokenNotFound)
		} else {
			response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("authentication failed: %v", err))
		}
		return
	}

	// Get platform implementation
	platform, err := h.registry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.Error(ctx, err, "platform not found", "provider", req.Provider)
		response.Error(c, errors.ErrPlatformNotSupported)
		return
	}

	// Get recent posts
	h.logger.Info(ctx, "getting recent posts", "provider", req.Provider, "user_id", req.UserID, "limit", req.Limit)
	posts, err := platform.GetRecentPosts(ctx, client, req.Limit, req.StartTime, req.EndTime)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get recent posts", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, err.Error())
		return
	}

	h.logger.Info(ctx, "recent posts retrieved successfully", "provider", req.Provider, "user_id", req.UserID, "count", len(posts))

	recentPostsResponse := types.GetRecentPostsResponse{
		Provider:   req.Provider,
		UserID:     req.UserID,
		ServerName: req.ServerName,
		Posts:      posts,
		Total:      len(posts),
	}
	response.Success(c, recentPostsResponse)
}

// BatchGetRecentPosts handles batch recent posts requests
// @Summary 批量获取最近发布的内容
// @Description 批量获取多个平台最近发布的内容列表，支持定时后驱
// @Tags 内容
// @Accept json
// @Produce json
// @Param request body types.BatchGetRecentPostsRequest true "批量获取最近发布内容请求参数"
// @Success 200 {object} types.APIResponse{data=types.BatchGetRecentPostsResponse} "获取成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/batch-recent-posts [post]
func (h *ShareHandler) BatchGetRecentPosts(c *gin.Context) {
	ctx := c.Request.Context()

	var req types.BatchGetRecentPostsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind batch recent posts request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Set default limits if not provided
	for i := range req.Platforms {
		if req.Platforms[i].Limit <= 0 {
			req.Platforms[i].Limit = 10
		}
	}

	// Get authenticated client with automatic token refresh
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var platformResults []types.PlatformPosts
	var totalPosts int
	var successCount int
	var errorCount int

	// Process each platform
	for _, platformReq := range req.Platforms {
		// Get authenticated client for this platform
		client, err := h.tokenManager.CreateAuthenticatedClient(ctx, req.UserID, platformReq.Provider, req.ServerName)
		if err != nil {
			h.logger.Error(ctx, err, "failed to create authenticated client", "provider", platformReq.Provider, "user_id", req.UserID)
			platformResults = append(platformResults, types.PlatformPosts{
				Provider:   platformReq.Provider,
				UserID:     req.UserID,
				ServerName: req.ServerName,
				Posts:      []types.Post{},
				Total:      0,
				Error:      fmt.Sprintf("authentication failed: %v", err),
			})
			errorCount++
			continue
		}

		// Get platform implementation
		platform, err := h.registry.GetPlatform(platformReq.Provider)
		if err != nil {
			h.logger.Error(ctx, err, "platform not found", "provider", platformReq.Provider)
			platformResults = append(platformResults, types.PlatformPosts{
				Provider:   platformReq.Provider,
				UserID:     req.UserID,
				ServerName: req.ServerName,
				Posts:      []types.Post{},
				Total:      0,
				Error:      "platform not supported",
			})
			errorCount++
			continue
		}

		// Get recent posts for this platform
		h.logger.Info(ctx, "getting recent posts", "provider", platformReq.Provider, "user_id", req.UserID, "limit", platformReq.Limit)
		posts, err := platform.GetRecentPosts(ctx, client, platformReq.Limit, req.StartTime, req.EndTime)
		if err != nil {
			h.logger.Error(ctx, err, "failed to get recent posts", "provider", platformReq.Provider, "user_id", req.UserID)
			platformResults = append(platformResults, types.PlatformPosts{
				Provider:   platformReq.Provider,
				UserID:     req.UserID,
				ServerName: req.ServerName,
				Posts:      []types.Post{},
				Total:      0,
				Error:      err.Error(),
			})
			errorCount++
			continue
		}

		// Success
		platformResults = append(platformResults, types.PlatformPosts{
			Provider:   platformReq.Provider,
			UserID:     req.UserID,
			ServerName: req.ServerName,
			Posts:      posts,
			Total:      len(posts),
		})
		totalPosts += len(posts)
		successCount++
	}

	h.logger.Info(ctx, "batch recent posts completed", "user_id", req.UserID, "success_count", successCount, "error_count", errorCount, "total_posts", totalPosts)

	batchResponse := types.BatchGetRecentPostsResponse{
		UserID:       req.UserID,
		ServerName:   req.ServerName,
		Platforms:    platformResults,
		TotalPosts:   totalPosts,
		SuccessCount: successCount,
		ErrorCount:   errorCount,
	}
	response.Success(c, batchResponse)
}
