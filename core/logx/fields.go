package logx

import "context"

type contextKey struct{}

var fieldsContextKey contextKey

// WithFields 返回具有给定字段的新上下文。
func WithFields(ctx context.Context, fields ...LogField) context.Context {
	if val := ctx.Value(fieldsContextKey); val != nil {
		if arr, ok := val.([]LogField); ok {
			return context.WithValue(ctx, fieldsContextKey, append(arr, fields...))
		}
	}

	return context.WithValue(ctx, fieldsContextKey, fields)
}
