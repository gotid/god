package logx

import (
	"context"
	"sync"
	"sync/atomic"
)

type contextKey struct{}

var (
	fieldsContextKey contextKey
	globalFields     atomic.Value
	globalFieldsLock sync.Mutex
)

// AddGlobalFields 增加全局字段。
func AddGlobalFields(fields ...LogField) {
	globalFieldsLock.Lock()
	defer globalFieldsLock.Unlock()

	old := globalFields.Load()
	if old == nil {
		globalFields.Store(append([]LogField(nil), fields...))
	} else {
		globalFields.Store(append(old.([]LogField), fields...))
	}
}

// ContextWithFields 返回具有给定字段的新上下文。
func ContextWithFields(ctx context.Context, fields ...LogField) context.Context {
	if val := ctx.Value(fieldsContextKey); val != nil {
		if arr, ok := val.([]LogField); ok {
			allFields := make([]LogField, 0, len(arr)+len(fields))
			allFields = append(allFields, arr...)
			allFields = append(allFields, fields...)
			return context.WithValue(ctx, fieldsContextKey, allFields)
		}
	}

	return context.WithValue(ctx, fieldsContextKey, fields)
}
