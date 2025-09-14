package response

import (
	"net/http"

	"social/internal/types"
	"social/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ResponseHandler 封装了HTTP响应处理
type ResponseHandler struct{}

// NewResponseHandler 创建新的响应处理器
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// Success 返回成功响应
func (r *ResponseHandler) Success(c *gin.Context, data interface{}) {
	requestID := r.getRequestID(c)

	response := types.APIResponse{
		Status:    "ok",
		Data:      data,
		RequestID: requestID,
	}

	c.JSON(http.StatusOK, response)
}

// SuccessWithMessage 返回带消息的成功响应
func (r *ResponseHandler) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	requestID := r.getRequestID(c)

	response := types.APIResponse{
		Status:    "ok",
		Message:   message,
		Data:      data,
		RequestID: requestID,
	}

	c.JSON(http.StatusOK, response)
}

// Error 返回错误响应
func (r *ResponseHandler) Error(c *gin.Context, appErr *errors.AppError) {
	requestID := r.getRequestID(c)

	response := types.ErrorResponse{
		Error:     appErr.Message,
		Code:      appErr.Code,
		RequestID: requestID,
	}

	c.JSON(appErr.Status, response)
}

// ErrorWithDetail 返回带详细信息的错误响应
func (r *ResponseHandler) ErrorWithDetail(c *gin.Context, appErr *errors.AppError, detail string) {
	requestID := r.getRequestID(c)

	response := types.ErrorResponse{
		Error:     appErr.Message,
		Code:      appErr.Code,
		RequestID: requestID,
	}

	// 在开发环境下可以包含详细信息
	if gin.Mode() == gin.DebugMode {
		response.Error = detail
	}

	c.JSON(appErr.Status, response)
}

// BadRequest 返回400错误
func (r *ResponseHandler) BadRequest(c *gin.Context, message string) {
	r.Error(c, &errors.AppError{
		Code:    errors.ErrInvalidRequest.Code,
		Message: message,
		Status:  http.StatusBadRequest,
	})
}

// Unauthorized 返回401错误
func (r *ResponseHandler) Unauthorized(c *gin.Context, message string) {
	r.Error(c, &errors.AppError{
		Code:    errors.ErrUnauthorized.Code,
		Message: message,
		Status:  http.StatusUnauthorized,
	})
}

// Forbidden 返回403错误
func (r *ResponseHandler) Forbidden(c *gin.Context, message string) {
	r.Error(c, &errors.AppError{
		Code:    errors.ErrForbidden.Code,
		Message: message,
		Status:  http.StatusForbidden,
	})
}

// NotFound 返回404错误
func (r *ResponseHandler) NotFound(c *gin.Context, message string) {
	r.Error(c, &errors.AppError{
		Code:    errors.ErrNotFound.Code,
		Message: message,
		Status:  http.StatusNotFound,
	})
}

// InternalServerError 返回500错误
func (r *ResponseHandler) InternalServerError(c *gin.Context, message string) {
	r.Error(c, &errors.AppError{
		Code:    errors.ErrInternalServer.Code,
		Message: message,
		Status:  http.StatusInternalServerError,
	})
}

// ServiceUnavailable 返回503错误
func (r *ResponseHandler) ServiceUnavailable(c *gin.Context, message string) {
	r.Error(c, &errors.AppError{
		Code:    errors.ErrServiceUnavailable.Code,
		Message: message,
		Status:  http.StatusServiceUnavailable,
	})
}

// Created 返回201创建成功响应
func (r *ResponseHandler) Created(c *gin.Context, data interface{}) {
	requestID := r.getRequestID(c)

	response := types.APIResponse{
		Status:    "created",
		Data:      data,
		RequestID: requestID,
	}

	c.JSON(http.StatusCreated, response)
}

// NoContent 返回204无内容响应
func (r *ResponseHandler) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Redirect 返回重定向响应
func (r *ResponseHandler) Redirect(c *gin.Context, url string) {
	c.Redirect(http.StatusFound, url)
}

// getRequestID 从上下文中获取请求ID
func (r *ResponseHandler) getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// 全局响应处理器实例
var DefaultResponseHandler = NewResponseHandler()

// 便捷函数，直接使用全局实例

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	DefaultResponseHandler.Success(c, data)
}

// SuccessWithMessage 返回带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	DefaultResponseHandler.SuccessWithMessage(c, message, data)
}

// Error 返回错误响应
func Error(c *gin.Context, appErr *errors.AppError) {
	DefaultResponseHandler.Error(c, appErr)
}

// ErrorWithDetail 返回带详细信息的错误响应
func ErrorWithDetail(c *gin.Context, appErr *errors.AppError, detail string) {
	DefaultResponseHandler.ErrorWithDetail(c, appErr, detail)
}

// BadRequest 返回400错误
func BadRequest(c *gin.Context, message string) {
	DefaultResponseHandler.BadRequest(c, message)
}

// Unauthorized 返回401错误
func Unauthorized(c *gin.Context, message string) {
	DefaultResponseHandler.Unauthorized(c, message)
}

// Forbidden 返回403错误
func Forbidden(c *gin.Context, message string) {
	DefaultResponseHandler.Forbidden(c, message)
}

// NotFound 返回404错误
func NotFound(c *gin.Context, message string) {
	DefaultResponseHandler.NotFound(c, message)
}

// InternalServerError 返回500错误
func InternalServerError(c *gin.Context, message string) {
	DefaultResponseHandler.InternalServerError(c, message)
}

// ServiceUnavailable 返回503错误
func ServiceUnavailable(c *gin.Context, message string) {
	DefaultResponseHandler.ServiceUnavailable(c, message)
}

// Created 返回201创建成功响应
func Created(c *gin.Context, data interface{}) {
	DefaultResponseHandler.Created(c, data)
}

// NoContent 返回204无内容响应
func NoContent(c *gin.Context) {
	DefaultResponseHandler.NoContent(c)
}

// Redirect 返回重定向响应
func Redirect(c *gin.Context, url string) {
	DefaultResponseHandler.Redirect(c, url)
}
