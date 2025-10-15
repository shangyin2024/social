package handlers

import (
	"context"
	"fmt"
	"net/url"
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

// AuthHandler handles OAuth authentication requests
type AuthHandler struct {
	config           *config.Config
	storage          storage.Storage
	logger           *logger.Logger
	platformRegistry *platforms.Registry
	tokenManager     *oauth.TokenManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config, storage storage.Storage, platformRegistry *platforms.Registry, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		config:           cfg,
		storage:          storage,
		logger:           logger,
		platformRegistry: platformRegistry,
		tokenManager:     oauth.NewTokenManager(cfg, storage, logger),
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
	ctx := c.Request.Context()

	var req types.StartAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind start auth request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Get OAuth config with server-specific configuration
	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, req.ServerName, req.RedirectURI)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", req.ServerName)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// Encode state with server name
	state, err := oauth.EncodeState(req.UserID, req.ServerName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to encode state")
		response.InternalServerError(c, "failed to generate state")
		return
	}

	// Create OAuth service
	oauthService := oauth.NewOAuthService(oauthConfig)

	// Generate auth URL
	usePKCE := req.Provider == "x" // Only X/Twitter uses PKCE
	authURL, verifier, err := oauthService.GenerateAuthURL(state, usePKCE)
	if err != nil {
		h.logger.Error(ctx, err, "failed to generate auth URL", "provider", req.Provider)
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
		h.logger.Info(ctx, "saving PKCE verifier", "state", state, "verifier_length", len(verifier), "verifier_preview", verifierPreview)

		if err := h.storage.SavePKCEVerifier(ctx, state, verifier); err != nil {
			h.logger.Error(ctx, err, "failed to save PKCE verifier", "state", state, "verifier_length", len(verifier))
			response.InternalServerError(c, "failed to save PKCE verifier")
			return
		}

		h.logger.Info(ctx, "PKCE verifier saved successfully", "state", state, "verifier_length", len(verifier))
	} else if usePKCE {
		h.logger.Error(ctx, errors.ErrInternalServer, "PKCE required but verifier is empty", "provider", req.Provider, "usePKCE", usePKCE, "verifier_empty", verifier == "")
		response.InternalServerError(c, "PKCE verifier generation failed")
		return
	}

	h.logger.Info(ctx, "OAuth flow initiated", "provider", req.Provider, "user_id", req.UserID, "server_name", req.ServerName)

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
	ctx := c.Request.Context()

	var req types.CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind callback request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Decode state
	statePayload, err := oauth.DecodeState(req.State)
	if err != nil {
		h.logger.Error(ctx, err, "failed to decode state")
		response.Error(c, errors.ErrInvalidState)
		return
	}

	h.logger.Info(ctx, "decoded state", "state", req.State, "state_payload user_id", statePayload.UserID, "state_payload server_name", statePayload.ServerName)

	// 使用请求中的服务内部用户ID，而不是state中的平台用户ID
	userID := req.UserID
	serverName := req.ServerName

	// 验证请求中的 server_name 与 state 中的 server_name 是否一致
	if req.ServerName != statePayload.ServerName {
		h.logger.Error(ctx, errors.ErrInvalidState, "server_name mismatch", "request_server", req.ServerName, "state_server", statePayload.ServerName)
		response.Error(c, errors.ErrInvalidState)
		return
	}

	// 记录平台用户ID用于日志和调试
	platformUserID := statePayload.UserID
	h.logger.Info(ctx, "processing OAuth callback", "service_user_id", userID, "platform_user_id", platformUserID, "server_name", serverName)

	// Get OAuth config with server-specific configuration
	// Use the redirect URI from the request or default callback URL
	// For token exchange, we need to use the exact same redirect_uri as used in authorization
	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = "https://test-pubproject.wondera.io/static/callback.html"
	}

	// For X platform, we need to ensure the redirect_uri matches exactly what was used in authorization
	// Remove any query parameters that might have been added during the callback
	if req.Provider == "x" {
		h.logger.Info(ctx, "processing X platform redirect_uri", "original_redirect_uri", redirectURI)
		// Parse the redirect URI to remove query parameters
		if parsedURL, err := url.Parse(redirectURI); err == nil {
			parsedURL.RawQuery = ""
			redirectURI = parsedURL.String()
			h.logger.Info(ctx, "cleaned X platform redirect_uri", "cleaned_redirect_uri", redirectURI)
		} else {
			h.logger.Error(ctx, err, "failed to parse redirect_uri", "redirect_uri", redirectURI)
		}
	}

	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, serverName, redirectURI)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", serverName)
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

		h.logger.Info(ctx, "getting PKCE verifier", "state", req.State, "provider", req.Provider)

		verifier, err = h.storage.GetAndDeletePKCEVerifier(ctx, req.State)
		if err != nil {
			h.logger.Error(ctx, err, "failed to get PKCE verifier", "provider", req.Provider, "state", req.State)
			response.ErrorWithDetail(c, errors.ErrInvalidState, "PKCE verifier not found or expired")
			return
		}

		verifierPreview := verifier
		if len(verifier) > 10 {
			verifierPreview = verifier[:10]
		}
		h.logger.Info(ctx, "PKCE verifier retrieved", "state", req.State, "verifier_length", len(verifier), "verifier_preview", verifierPreview)
	}

	// Exchange authorization code for token
	token, err := oauthService.ExchangeCode(ctx, req.Code, verifier)
	if err != nil {
		h.logger.Error(ctx, err, "token exchange failed", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("token exchange failed: %v", err))
		return
	}

	// Save token to storage
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	h.logger.Info(ctx, "attempting to save token", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID, "server_name", serverName, "token_type", token.TokenType)

	if err := h.storage.SaveToken(ctx, userID, req.Provider, serverName, token); err != nil {
		h.logger.Error(ctx, err, "failed to save token", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID, "server_name", serverName)
		response.ErrorWithDetail(c, errors.ErrInternalServer, "failed to save token")
		return
	}

	// Verify token was saved successfully by trying to retrieve it
	ctx2, cancel2 := context.WithTimeout(ctx, 3*time.Second)
	defer cancel2()

	savedToken, err := h.storage.GetToken(ctx2, userID, req.Provider, serverName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to verify token save", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID, "server_name", serverName)
		response.ErrorWithDetail(c, errors.ErrInternalServer, "token save verification failed")
		return
	}

	if savedToken.AccessToken != token.AccessToken {
		h.logger.Error(ctx, errors.ErrInternalServer, "token save verification failed - access token mismatch", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID, "server_name", serverName)
		response.ErrorWithDetail(c, errors.ErrInternalServer, "token save verification failed")
		return
	}

	h.logger.Info(ctx, "token saved and verified successfully", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID, "server_name", serverName)

	// 获取平台实例并调用平台特定的OAuth回调处理
	platformInstance, err := h.platformRegistry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get platform", "provider", req.Provider)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// 调用平台特定的OAuth回调处理（用于平台特定的后处理）
	err = platformInstance.HandleOAuthCallback(ctx, req.Code, req.State)
	if err != nil {
		h.logger.Error(ctx, err, "platform OAuth callback failed", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID)
		// 这里不返回错误，因为token已经保存成功，平台特定的处理失败不应该影响整个流程
		h.logger.Info(ctx, "platform callback failed but token saved successfully", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID)
	}

	h.logger.Info(ctx, "OAuth flow completed successfully", "provider", req.Provider, "service_user_id", userID, "platform_user_id", platformUserID)

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

// 查询是否授权
// @Summary 查询是否授权
// @Description 查询指定用户是否授权指定平台
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body types.IsAuthorizedRequest true "查询是否授权请求参数"
// @Success 200 {object} types.APIResponse{data=types.IsAuthorizedResponse} "查询成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /auth/is-authorized [post]
func (h *AuthHandler) IsAuthorized(c *gin.Context) {
	ctx := c.Request.Context()

	var req types.IsAuthorizedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind is authorized request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Check if token is valid (without refreshing)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	isValid, err := h.tokenManager.IsTokenValid(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to check token validity", "provider", req.Provider, "user_id", req.UserID)
		response.Error(c, errors.ErrInternalServer)
		return
	}

	if !isValid {
		h.logger.Error(ctx, errors.ErrTokenExpired, "token not valid", "provider", req.Provider, "user_id", req.UserID)
		response.Error(c, errors.ErrTokenExpired)
		return
	}

	response.Success(c, types.IsAuthorizedResponse{
		IsAuthorized: true,
	})
}

// GetUserInfo retrieves user information from the platform
// @Summary 获取用户信息
// @Description 获取指定平台用户的详细信息
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body types.GetUserInfoRequest true "获取用户信息请求参数"
// @Success 200 {object} types.APIResponse{data=types.GetUserInfoResponse} "获取用户信息成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "未授权或token过期"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /auth/user-info [post]
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()

	var req types.GetUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind get user info request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Check if token is valid and get token
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token, err := h.tokenManager.GetValidToken(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get valid token", "provider", req.Provider, "user_id", req.UserID)
		response.Error(c, errors.ErrTokenExpired)
		return
	}

	// Get platform instance
	platformInstance, err := h.platformRegistry.GetPlatform(req.Provider)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get platform", "provider", req.Provider)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	// Create OAuth service to get HTTP client with token
	oauthConfig, err := h.config.GetServerOAuthConfig(req.Provider, req.ServerName, "")
	if err != nil {
		h.logger.Error(ctx, err, "failed to get OAuth config", "provider", req.Provider, "server_name", req.ServerName)
		response.ErrorWithDetail(c, errors.ErrInvalidProvider, err.Error())
		return
	}

	oauthService := oauth.NewOAuthService(oauthConfig)
	client := oauthService.CreateClient(ctx, token)

	// Get user info from platform
	userInfo, err := platformInstance.GetUserInfo(ctx, client)
	if err != nil {
		h.logger.Error(ctx, err, "failed to get user info", "provider", req.Provider, "user_id", req.UserID)
		response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("failed to get user info: %v", err))
		return
	}

	h.logger.Info(ctx, "user info retrieved successfully", "provider", req.Provider, "user_id", req.UserID, "platform_user_id", userInfo.ID)

	userInfoResponse := types.GetUserInfoResponse{
		Provider:   req.Provider,
		UserID:     req.UserID,
		ServerName: req.ServerName,
		UserInfo:   userInfo,
	}

	response.Success(c, userInfoResponse)
}

// RefreshToken handles manual token refresh requests
// @Summary 手动刷新token
// @Description 客户端主动刷新指定平台的访问token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body types.RefreshTokenRequest true "刷新token请求参数"
// @Success 200 {object} types.APIResponse{data=types.RefreshTokenResponse} "token刷新成功"
// @Failure 400 {object} types.ErrorResponse "请求参数错误"
// @Failure 401 {object} types.ErrorResponse "token不存在或刷新失败"
// @Failure 500 {object} types.ErrorResponse "服务器内部错误"
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()

	var req types.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx, err, "failed to bind refresh token request")
		response.BadRequest(c, "invalid request format")
		return
	}

	// Force refresh the token
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	newToken, err := h.tokenManager.ForceRefreshToken(ctx, req.UserID, req.Provider, req.ServerName)
	if err != nil {
		h.logger.Error(ctx, err, "failed to refresh token", "provider", req.Provider, "user_id", req.UserID)
		if err.Error() == "OAuth token not found" {
			response.ErrorWithDetail(c, errors.ErrTokenNotFound, "Token not found. Please re-authorize your account.")
		} else {
			response.ErrorWithDetail(c, errors.ErrInternalServer, fmt.Sprintf("token refresh failed: %v", err))
		}
		return
	}

	// Calculate timestamps
	var expiresAt int64
	if !newToken.Expiry.IsZero() {
		expiresAt = newToken.Expiry.Unix()
	}
	refreshedAt := time.Now().Unix()

	h.logger.Info(ctx, "token refreshed successfully", "provider", req.Provider, "user_id", req.UserID, "server_name", req.ServerName)

	refreshResponse := types.RefreshTokenResponse{
		Provider:    req.Provider,
		UserID:      req.UserID,
		ServerName:  req.ServerName,
		ExpiresAt:   expiresAt,
		RefreshedAt: refreshedAt,
		Message:     fmt.Sprintf("Token refreshed successfully for user %s on %s platform", req.UserID, req.Provider),
	}

	response.SuccessWithMessage(c, "Token refreshed successfully", refreshResponse)
}
