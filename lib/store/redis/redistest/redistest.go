package redistest

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/store/redis"
	"time"
)

// CreateRedis 返回一个用于测试的进程中的 redis.Redis。
func CreateRedis() (r *redis.Redis, clean func(), err error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}

	clean = func() {
		ch := make(chan lang.PlaceholderType)

		go func() {
			mr.Close()
			close(ch)
		}()

		select {
		case <-ch:
		case <-time.After(time.Second):
		}
	}
	return redis.New(mr.Addr()), clean, nil
}
