package handlers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"social/internal/config"
	"social/internal/oauth"
	"social/internal/platforms"
	"social/internal/storage"
	"social/internal/types"
	ctxutil "social/pkg/context"
	"social/pkg/errors"
	"social/pkg/logger"
	"social/pkg/response"
	"social/pkg/validator"
)

// ShareHandler handles content sharing requests
type ShareHandler struct {
	config   *config.Config
	storage  storage.Storage
	registry *platforms.Registry
	logger   *logger.Logger
}

// NewShareHandler creates a new share handler
func NewShareHandler(cfg *config.Config, storage storage.Storage, registry *platforms.Registry, logger *logger.Logger) *ShareHandler {
	return &ShareHandler{
		config:   cfg,
		storage:  storage,
		registry: registry,
		logger:   logger,
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
	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind share request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// 使用自定义验证器进行额外验证
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.GetValidationErrors(err)
		h.logger.LogError(ctx, err, "validation failed", "errors", validationErrors)
		response.BadRequest(c, "validation failed")
		return
	}

	// Validate provider
	if !h.config.IsProviderSupported(req.Provider) {
		h.logger.LogError(ctx, errors.ErrInvalidProvider, "unsupported provider", "provider", req.Provider)
		response.Error(c, errors.ErrInvalidProvider)
		return
	}

	// Get token
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token, err := h.storage.GetToken(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.LogError(ctx, err, "token not found", "provider", req.Provider, "user_id", req.UserID)
		response.Error(c, errors.ErrTokenNotFound)
		return
	}

	// Get OAuth config with server-specific configuration
	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, req.ServerName, "")
	if err != nil {
		h.logger.LogError(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", req.ServerName)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// Create OAuth service and client
	oauthService := oauth.NewOAuthService(oauthConfig)
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	client := oauthService.CreateClient(ctx, token)

	// Get platform implementation
	platform, err := h.registry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.LogError(ctx, err, "platform not found", "provider", req.Provider)
		response.Error(c, errors.ErrPlatformNotSupported)
		return
	}

	// Share content
	h.logger.LogInfo(ctx, "sharing content", "provider", req.Provider, "user_id", req.UserID)
	if err := platform.Share(ctx, client, &req); err != nil {
		h.logger.LogError(ctx, err, "failed to share content", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, err.Error())
		return
	}

	h.logger.LogInfo(ctx, "content shared successfully", "provider", req.Provider, "user_id", req.UserID)

	// Get media ID from context if available (set by platform implementation)
	var mediaID string
	if id := ctx.Value("tweet_id"); id != nil {
		if tweetID, ok := id.(string); ok {
			mediaID = tweetID
		}
	}

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
	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.StatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind stats request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// 使用自定义验证器进行额外验证
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.GetValidationErrors(err)
		h.logger.LogError(ctx, err, "validation failed", "errors", validationErrors)
		response.BadRequest(c, "validation failed")
		return
	}

	// Validate provider
	if !h.config.IsProviderSupported(req.Provider) {
		h.logger.LogError(ctx, errors.ErrInvalidProvider, "unsupported provider", "provider", req.Provider)
		response.Error(c, errors.ErrInvalidProvider)
		return
	}

	// Get token
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token, err := h.storage.GetToken(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.LogError(ctx, err, "token not found", "provider", req.Provider, "user_id", req.UserID)
		response.Error(c, errors.ErrTokenNotFound)
		return
	}

	// Get OAuth config with server-specific configuration
	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, req.ServerName, "")
	if err != nil {
		h.logger.LogError(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", req.ServerName)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// Create OAuth service and client
	oauthService := oauth.NewOAuthService(oauthConfig)
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	client := oauthService.CreateClient(ctx, token)

	// Get platform implementation
	platform, err := h.registry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.LogError(ctx, err, "platform not found", "provider", req.Provider)
		response.Error(c, errors.ErrPlatformNotSupported)
		return
	}

	// Get statistics
	h.logger.LogInfo(ctx, "getting statistics", "provider", req.Provider, "user_id", req.UserID, "media_id", req.MediaID)
	stats, err := platform.GetStats(ctx, client, req.MediaID)
	if err != nil {
		h.logger.LogError(ctx, err, "failed to get statistics", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, err.Error())
		return
	}

	h.logger.LogInfo(ctx, "statistics retrieved successfully", "provider", req.Provider, "user_id", req.UserID)

	statsResponse := types.StatsResponse{
		Provider:   req.Provider,
		UserID:     req.UserID,
		ServerName: req.ServerName,
		MediaID:    req.MediaID,
		Stats:      stats,
	}
	response.Success(c, statsResponse)
}

// CreatePost handles RESTful post creation
// @Summary 创建社交媒体帖子
// @Description 使用RESTful API创建新的社交媒体帖子
// @Tags 内容管理
// @Accept json
// @Produce json
// @Param platform path string true "平台名称" example:"x"
// @Param request body types.ShareRequest true "创建帖子请求参数"
// @Success 201 {object} types.APIResponse{data=types.ShareResponse} "帖子创建成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/{platform}/posts [post]
func (h *ShareHandler) CreatePost(c *gin.Context) {
	// Extract platform from URL path
	platform := c.Param("platform")
	if platform == "" {
		response.BadRequest(c, "platform parameter is required")
		return
	}

	// Validate platform
	validPlatforms := []string{"youtube", "x", "facebook", "tiktok", "instagram"}
	if !contains(validPlatforms, platform) {
		response.BadRequest(c, "invalid platform")
		return
	}

	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind create post request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Set provider from URL
	req.Provider = platform

	// Use custom validator for additional validation
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.GetValidationErrors(err)
		h.logger.LogError(ctx, err, "validation failed", "errors", validationErrors)
		response.BadRequest(c, "validation failed")
		return
	}

	// Call the existing Share method
	h.Share(c)
}

// GetPostStats handles RESTful post statistics retrieval
// @Summary 获取帖子统计信息
// @Description 使用RESTful API获取特定帖子的统计信息
// @Tags 统计分析
// @Accept json
// @Produce json
// @Param platform path string true "平台名称" example:"x"
// @Param post_id path string true "帖子ID" example:"1234567890"
// @Param request body types.StatsRequest true "统计请求参数"
// @Success 200 {object} types.APIResponse{data=types.StatsResponse} "统计信息获取成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/{platform}/posts/{post_id}/stats [get]
func (h *ShareHandler) GetPostStats(c *gin.Context) {
	// Extract platform and post_id from URL path
	platform := c.Param("platform")
	postID := c.Param("post_id")

	if platform == "" || postID == "" {
		response.BadRequest(c, "platform and post_id parameters are required")
		return
	}

	// Validate platform
	validPlatforms := []string{"youtube", "x", "facebook", "tiktok", "instagram"}
	if !contains(validPlatforms, platform) {
		response.BadRequest(c, "invalid platform")
		return
	}

	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.StatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind get post stats request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Set provider and media_id from URL
	req.Provider = platform
	req.MediaID = postID

	// Use custom validator for additional validation
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.GetValidationErrors(err)
		h.logger.LogError(ctx, err, "validation failed", "errors", validationErrors)
		response.BadRequest(c, "validation failed")
		return
	}

	// Call the existing GetStats method
	h.GetStats(c)
}

// GetUserStats handles RESTful user statistics retrieval
// @Summary 获取用户统计信息
// @Description 使用RESTful API获取用户的整体统计信息
// @Tags 统计分析
// @Accept json
// @Produce json
// @Param platform path string true "平台名称" example:"x"
// @Param user_id path string true "用户ID" example:"user123"
// @Param request body types.StatsRequest true "统计请求参数"
// @Success 200 {object} types.APIResponse{data=types.StatsResponse} "用户统计信息获取成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /api/{platform}/users/{user_id}/stats [get]
func (h *ShareHandler) GetUserStats(c *gin.Context) {
	// Extract platform and user_id from URL path
	platform := c.Param("platform")
	userID := c.Param("user_id")

	if platform == "" || userID == "" {
		response.BadRequest(c, "platform and user_id parameters are required")
		return
	}

	// Validate platform
	validPlatforms := []string{"youtube", "x", "facebook", "tiktok", "instagram"}
	if !contains(validPlatforms, platform) {
		response.BadRequest(c, "invalid platform")
		return
	}

	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.StatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind get user stats request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Set provider and user_id from URL
	req.Provider = platform
	req.UserID = userID

	// Use custom validator for additional validation
	if err := validator.Validate(&req); err != nil {
		validationErrors := validator.GetValidationErrors(err)
		h.logger.LogError(ctx, err, "validation failed", "errors", validationErrors)
		response.BadRequest(c, "validation failed")
		return
	}

	// For user stats, we might want to aggregate multiple posts
	// For now, we'll use the existing GetStats method but with empty MediaID
	req.MediaID = ""

	// Call the existing GetStats method
	h.GetStats(c)
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
