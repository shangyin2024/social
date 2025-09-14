package context

import "context"

// Key 用于 context 值的键类型
type Key string

const (
	// RequestIDKey 请求ID的context键
	RequestIDKey Key = "request_id"
)

// WithRequestID 将请求ID添加到context中
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID 从context中获取请求ID
func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	return requestID, ok
}
