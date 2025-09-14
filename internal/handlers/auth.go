package handlers

import (
	"context"
	"fmt"
	"net/url"
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

// AuthHandler handles OAuth authentication requests
type AuthHandler struct {
	config           *config.Config
	storage          storage.Storage
	logger           *logger.Logger
	platformRegistry *platforms.Registry
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config, storage storage.Storage, platformRegistry *platforms.Registry, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		config:           cfg,
		storage:          storage,
		logger:           logger,
		platformRegistry: platformRegistry,
	}
}

// StartAuth initiates OAuth flow
// @Summary 开始OAuth授权流程
// @Description 启动指定平台的OAuth授权流程，返回授权URL
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body types.StartAuthRequest true "授权请求参数"
// @Success 200 {object} types.APIResponse{data=types.StartAuthResponse} "授权URL生成成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /auth/start [post]
func (h *AuthHandler) StartAuth(c *gin.Context) {
	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.StartAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind start auth request")
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

	// Get OAuth config with server-specific configuration
	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, req.ServerName, req.RedirectURI)
	if err != nil {
		h.logger.LogError(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", req.ServerName)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// Encode state with server name
	state, err := oauth.EncodeState(req.UserID, req.ServerName)
	if err != nil {
		h.logger.LogError(ctx, err, "failed to encode state")
		response.InternalServerError(c, "failed to generate state")
		return
	}

	// Create OAuth service
	oauthService := oauth.NewOAuthService(oauthConfig)

	// Generate auth URL
	usePKCE := req.Provider == "x" // Only X/Twitter uses PKCE
	authURL, verifier, err := oauthService.GenerateAuthURL(state, usePKCE)
	if err != nil {
		h.logger.LogError(ctx, err, "failed to generate auth URL", "provider", req.Provider)
		response.InternalServerError(c, "failed to generate auth URL")
		return
	}

	// Store PKCE verifier if needed
	if usePKCE && verifier != "" {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		verifierPreview := verifier
		if len(verifier) > 10 {
			verifierPreview = verifier[:10]
		}
		h.logger.LogInfo(ctx, "saving PKCE verifier", "state", state, "verifier_length", len(verifier), "verifier_preview", verifierPreview)

		if err := h.storage.SavePKCEVerifier(ctx, state, verifier); err != nil {
			h.logger.LogError(ctx, err, "failed to save PKCE verifier", "state", state, "verifier_length", len(verifier))
			response.InternalServerError(c, "failed to save PKCE verifier")
			return
		}

		h.logger.LogInfo(ctx, "PKCE verifier saved successfully", "state", state, "verifier_length", len(verifier))
	} else if usePKCE {
		h.logger.LogError(ctx, errors.ErrInternalServer, "PKCE required but verifier is empty", "provider", req.Provider, "usePKCE", usePKCE, "verifier_empty", verifier == "")
		response.InternalServerError(c, "PKCE verifier generation failed")
		return
	}

	h.logger.LogInfo(ctx, "OAuth flow initiated", "provider", req.Provider, "user_id", req.UserID, "server_name", req.ServerName)

	// 返回授权 URL，让前端处理重定向
	authResponse := types.StartAuthResponse{
		AuthURL:    authURL,
		Provider:   req.Provider,
		UserID:     req.UserID,
		ServerName: req.ServerName,
	}
	response.Success(c, authResponse)
}

// Callback handles OAuth callback
// @Summary 处理OAuth回调
// @Description 前端收到第三方平台OAuth回调后，调用此接口处理授权码交换和token保存
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body types.CallbackRequest true "回调请求参数" example:"{\"provider\":\"x\",\"state\":\"encoded_state_string\",\"code\":\"authorization_code\"}"
// @Success 200 {object} types.APIResponse{data=types.CallbackResponse} "OAuth callback completed"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /auth/callback [post]
func (h *AuthHandler) Callback(c *gin.Context) {
	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind callback request")
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

	// Decode state
	statePayload, err := oauth.DecodeState(req.State)
	if err != nil {
		h.logger.LogError(ctx, err, "failed to decode state")
		response.Error(c, errors.ErrInvalidState)
		return
	}

	userID := statePayload.UserID
	serverName := statePayload.ServerName
	if userID == "" {
		h.logger.LogError(ctx, errors.ErrInvalidState, "state missing user_id")
		response.Error(c, errors.ErrInvalidState)
		return
	}
	if serverName == "" {
		h.logger.LogError(ctx, errors.ErrInvalidState, "state missing server_name")
		response.Error(c, errors.ErrInvalidState)
		return
	}

	// 验证请求中的 server_name 与 state 中的 server_name 是否一致
	if req.ServerName != serverName {
		h.logger.LogError(ctx, errors.ErrInvalidState, "server_name mismatch", "request_server", req.ServerName, "state_server", serverName)
		response.Error(c, errors.ErrInvalidState)
		return
	}

	// Get OAuth config with server-specific configuration
	// Use the redirect URI from the request or default callback URL
	// For token exchange, we need to use the exact same redirect_uri as used in authorization
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = "https://test-pubproject.wondera.ai/static/callback.html"
	}

	// For X platform, we need to ensure the redirect_uri matches exactly what was used in authorization
	// Remove any query parameters that might have been added during the callback
	if req.Provider == "x" {
		h.logger.LogInfo(ctx, "processing X platform redirect_uri", "original_redirect_uri", redirectURI)
		// Parse the redirect URI to remove query parameters
		if parsedURL, err := url.Parse(redirectURI); err == nil {
			parsedURL.RawQuery = ""
			redirectURI = parsedURL.String()
			h.logger.LogInfo(ctx, "cleaned X platform redirect_uri", "cleaned_redirect_uri", redirectURI)
		} else {
			h.logger.LogError(ctx, err, "failed to parse redirect_uri", "redirect_uri", redirectURI)
		}
	}

	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, serverName, redirectURI)
	if err != nil {
		h.logger.LogError(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", serverName)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// Create OAuth service
	oauthService := oauth.NewOAuthService(oauthConfig)

	// Get PKCE verifier if needed (for X platform)
	var verifier string
	if req.Provider == "x" {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		h.logger.LogInfo(ctx, "getting PKCE verifier", "state", req.State, "provider", req.Provider)

		verifier, err = h.storage.GetAndDeletePKCEVerifier(ctx, req.State)
		if err != nil {
			h.logger.LogError(ctx, err, "failed to get PKCE verifier", "provider", req.Provider, "state", req.State)
			response.ErrorWithDetail(c, errors.ErrInvalidState, "PKCE verifier not found or expired")
			return
		}

		verifierPreview := verifier
		if len(verifier) > 10 {
			verifierPreview = verifier[:10]
		}
		h.logger.LogInfo(ctx, "PKCE verifier retrieved", "state", req.State, "verifier_length", len(verifier), "verifier_preview", verifierPreview)
	}

	// Exchange authorization code for token
	token, err := oauthService.ExchangeCode(ctx, req.Code, verifier)
	if err != nil {
		h.logger.LogError(ctx, err, "token exchange failed", "provider", req.Provider, "user_id", userID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("token exchange failed: %v", err))
		return
	}

	// Save token to storage
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	h.logger.LogInfo(ctx, "attempting to save token", "provider", req.Provider, "user_id", userID, "server_name", serverName, "token_type", token.TokenType)

	if err := h.storage.SaveToken(ctx, userID, req.Provider, serverName, token); err != nil {
		h.logger.LogError(ctx, err, "failed to save token", "provider", req.Provider, "user_id", userID, "server_name", serverName)
		response.ErrorWithDetail(c, errors.ErrInternalServer, "failed to save token")
		return
	}

	h.logger.LogInfo(ctx, "token saved successfully", "provider", req.Provider, "user_id", userID, "server_name", serverName)

	// 获取平台实例并调用平台特定的OAuth回调处理
	platformInstance, err := h.platformRegistry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.LogError(ctx, err, "failed to get platform", "provider", req.Provider)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// 调用平台特定的OAuth回调处理（用于平台特定的后处理）
	err = platformInstance.HandleOAuthCallback(ctx, req.Code, req.State)
	if err != nil {
		h.logger.LogError(ctx, err, "platform OAuth callback failed", "provider", req.Provider, "user_id", userID)
		// 这里不返回错误，因为token已经保存成功，平台特定的处理失败不应该影响整个流程
		h.logger.LogInfo(ctx, "platform callback failed but token saved successfully", "provider", req.Provider, "user_id", userID)
	}

	h.logger.LogInfo(ctx, "OAuth flow completed successfully", "provider", req.Provider, "user_id", userID)

	// 计算时间戳
	var expiresAt int64
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Unix()
	}
	referAt := time.Now().Unix()

	callbackResponse := types.CallbackResponse{
		Provider:   req.Provider,
		UserID:     userID,
		ServerName: serverName,
		ExpiresAt:  expiresAt,
		ReferAt:    referAt,
		Message:    fmt.Sprintf("OAuth callback completed for user %s provider %s. You may close this window.", userID, req.Provider),
	}
	response.SuccessWithMessage(c, "OAuth callback completed successfully", callbackResponse)
}

// RefreshToken handles token refresh requests
// @Summary 刷新访问令牌
// @Description 使用refresh token获取新的access token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body types.RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} types.APIResponse{data=types.RefreshTokenResponse} "刷新成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	requestID := uuid.New().String()
	ctx := ctxutil.WithRequestID(c.Request.Context(), requestID)

	var req types.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.LogError(ctx, err, "failed to bind refresh token request")
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

	// Get current token from storage
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	currentToken, err := h.storage.GetToken(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.LogError(ctx, err, "token not found", "provider", req.Provider, "user_id", req.UserID, "server_name", req.ServerName)
		response.Error(c, errors.ErrTokenNotFound)
		return
	}

	// Check if refresh token exists
	if currentToken.RefreshToken == "" {
		h.logger.LogError(ctx, errors.ErrTokenNotFound, "refresh token not found", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrTokenNotFound, "refresh token not available")
		return
	}

	// Get OAuth config
	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, req.ServerName, "")
	if err != nil {
		h.logger.LogError(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", req.ServerName)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// Create OAuth service
	oauthService := oauth.NewOAuthService(oauthConfig)

	// Refresh token
	h.logger.LogInfo(ctx, "refreshing token", "provider", req.Provider, "user_id", req.UserID, "server_name", req.ServerName)

	newToken, err := oauthService.RefreshToken(ctx, currentToken.RefreshToken)
	if err != nil {
		h.logger.LogError(ctx, err, "token refresh failed", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("token refresh failed: %v", err))
		return
	}

	// Save new token to storage
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.storage.SaveToken(ctx, req.UserID, req.Provider, req.ServerName, newToken); err != nil {
		h.logger.LogError(ctx, err, "failed to save refreshed token", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, "failed to save refreshed token")
		return
	}

	h.logger.LogInfo(ctx, "token refreshed successfully", "provider", req.Provider, "user_id", req.UserID, "server_name", req.ServerName)

	// 计算时间戳
	var expiresAt int64
	if !newToken.Expiry.IsZero() {
		expiresAt = newToken.Expiry.Unix()
	}
	referAt := time.Now().Unix()

	refreshResponse := types.RefreshTokenResponse{
		Provider:   req.Provider,
		UserID:     req.UserID,
		ServerName: req.ServerName,
		ExpiresAt:  expiresAt,
		ReferAt:    referAt,
	}
	response.SuccessWithMessage(c, "token refreshed successfully", refreshResponse)
}
