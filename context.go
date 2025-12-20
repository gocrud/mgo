package mgo

import (
	"context"
	"time"
)

// ==================== Context 管理 ====================

// defaultContext 默认上下文（context.Background）
var defaultContext = context.Background()

// getContext 获取上下文，如果为 nil 则返回默认上下文
func getContext(ctx context.Context) context.Context {
	if ctx == nil {
		return defaultContext
	}
	return ctx
}

// ==================== Context 辅助函数 ====================

// WithTimeout 创建带超时的上下文
//
// 示例：
//
//	ctx := mgo.WithTimeout(5 * time.Second)
//	users.Find().Ctx(ctx).All()
func WithTimeout(timeout time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	return ctx
}

// WithDeadline 创建带截止时间的上下文
//
// 示例：
//
//	deadline := time.Now().Add(5 * time.Second)
//	ctx := mgo.WithDeadline(deadline)
//	users.Find().Ctx(ctx).All()
func WithDeadline(deadline time.Time) context.Context {
	ctx, _ := context.WithDeadline(context.Background(), deadline)
	return ctx
}

// WithCancel 创建可取消的上下文
//
// 示例：
//
//	ctx, cancel := mgo.WithCancel()
//	defer cancel()
//	users.Find().Ctx(ctx).All()
func WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// ==================== Context 键值对 ====================

type contextKey string

const (
	// contextKeyUserID 用户 ID 上下文键
	contextKeyUserID contextKey = "user_id"

	// contextKeyTraceID 追踪 ID 上下文键
	contextKeyTraceID contextKey = "trace_id"

	// contextKeyRequestID 请求 ID 上下文键
	contextKeyRequestID contextKey = "request_id"
)

// WithUserID 将用户 ID 添加到上下文
//
// 示例：
//
//	ctx := mgo.WithUserID(context.Background(), "user123")
//	users.Find().Ctx(ctx).All()
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// UserIDFromContext 从上下文获取用户 ID
//
// 示例：
//
//	userID, ok := mgo.UserIDFromContext(ctx)
//	if ok {
//	    fmt.Println("User ID:", userID)
//	}
func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(contextKeyUserID).(string)
	return userID, ok
}

// WithTraceID 将追踪 ID 添加到上下文
//
// 示例：
//
//	ctx := mgo.WithTraceID(context.Background(), "trace123")
//	users.Find().Ctx(ctx).All()
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, contextKeyTraceID, traceID)
}

// TraceIDFromContext 从上下文获取追踪 ID
//
// 示例：
//
//	traceID, ok := mgo.TraceIDFromContext(ctx)
//	if ok {
//	    fmt.Println("Trace ID:", traceID)
//	}
func TraceIDFromContext(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(contextKeyTraceID).(string)
	return traceID, ok
}

// WithRequestID 将请求 ID 添加到上下文
//
// 示例：
//
//	ctx := mgo.WithRequestID(context.Background(), "req123")
//	users.Find().Ctx(ctx).All()
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, requestID)
}

// RequestIDFromContext 从上下文获取请求 ID
//
// 示例：
//
//	requestID, ok := mgo.RequestIDFromContext(ctx)
//	if ok {
//	    fmt.Println("Request ID:", requestID)
//	}
func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(contextKeyRequestID).(string)
	return requestID, ok
}
