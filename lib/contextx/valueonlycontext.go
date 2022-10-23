package contextx

import (
	"context"
	"time"
)

type valueOnlyContext struct {
	context.Context
}

func (valueOnlyContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (valueOnlyContext) Done() <-chan struct{} {
	return nil
}

func (valueOnlyContext) Err() error {
	return nil
}

// ValueOnlyFrom 返回除了 deadline 和 错误控制之外的所有值。
func ValueOnlyFrom(ctx context.Context) context.Context {
	return valueOnlyContext{
		Context: ctx,
	}
}
