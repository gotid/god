package queue

import "github.com/gotid/god/lib/errorx"

// MultiPusher 是一个推送消息给多个底层推手的推手 Pusher。
type MultiPusher struct {
	name    string
	pushers []Pusher
}

// NewMultiPusher 返回一个 MultiPusher。
func NewMultiPusher(pushers []Pusher) Pusher {
	return &MultiPusher{
		name:    generateName(pushers),
		pushers: pushers,
	}
}

func (mp *MultiPusher) Name() string {
	return mp.name
}

func (mp *MultiPusher) Push(message string) error {
	var batchError errorx.BatchError

	for _, each := range mp.pushers {
		if err := each.Push(message); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Err()
}
