package queue

import (
	"errors"
	"github.com/gotid/god/lib/logx"
	"sync/atomic"
)

// ErrNoAvailablePusher 表示没有可用的 Pusher。
var ErrNoAvailablePusher = errors.New("没有可用的推手")

// BalancedPusher 基于 round-robin 算法，推送消息给多个推手。
type BalancedPusher struct {
	name    string
	pushers []Pusher
	index   uint64
}

// NewBalancedPusher 返回一个 BalancedPusher。
func NewBalancedPusher(pushers []Pusher) Pusher {
	return &BalancedPusher{
		name:    generateName(pushers),
		pushers: pushers,
	}
}

func (bp *BalancedPusher) Name() string {
	return bp.name
}

func (bp *BalancedPusher) Push(message string) error {
	size := len(bp.pushers)

	for i := 0; i < size; i++ {
		index := atomic.AddUint64(&bp.index, 1) % uint64(size)
		target := bp.pushers[index]

		if err := target.Push(message); err != nil {
			logx.Error(err)
		} else {
			return nil
		}
	}

	return ErrNoAvailablePusher
}
