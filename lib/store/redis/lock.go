package redis

import (
	"context"
	red "github.com/go-redis/redis/v8"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stringx"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	randomLen       = 16
	millisPerSecond = 1000
	tolerance       = 500 // 毫秒
	lockCommand     = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
    return "OK"
else
    return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
end`
	delCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`
)

// Lock 是一把 redis 锁。
type Lock struct {
	rds     *Redis
	seconds uint32
	key     string
	id      string
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewLock 返回一个 Lock 实例。
func NewLock(rds *Redis, key string) *Lock {
	return &Lock{
		rds: rds,
		key: key,
		id:  stringx.Randn(randomLen),
	}
}

// Acquire 获取 redis 锁。
func (l *Lock) Acquire() (bool, error) {
	return l.AcquireCtx(context.Background())
}

// AcquireCtx 获取具有给定上下文的 redis 锁。
func (l *Lock) AcquireCtx(ctx context.Context) (bool, error) {
	seconds := atomic.LoadUint32(&l.seconds)
	resp, err := l.rds.EvalCtx(ctx, lockCommand, []string{l.key}, []string{
		l.id, strconv.Itoa(int(seconds)*millisPerSecond + tolerance),
	})
	if err == red.Nil {
		return false, nil
	} else if err != nil {
		logx.Errorf("为键 %s 获取 redis 锁时发生错误：%s", l.key, err.Error())
		return false, err
	} else if resp == nil {
		return false, nil
	}

	if reply, ok := resp.(string); ok && reply == "OK" {
		return true, nil
	}

	logx.Errorf("为键 %s 获取 redis 锁时返回未知响应：%s", l.key, resp)
	return false, nil
}

// Release 释放 redis 锁。
func (l *Lock) Release() (bool, error) {
	return l.ReleaseCtx(context.Background())
}

// ReleaseCtx 释放给定上下文的 redis 锁。
func (l *Lock) ReleaseCtx(ctx context.Context) (bool, error) {
	resp, err := l.rds.EvalCtx(ctx, delCommand, []string{l.key}, []string{l.id})
	if err != nil {
		return false, err
	}

	reply, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return reply == 1, nil
}

// SetExpire 设置过期时间
func (l *Lock) SetExpire(seconds int) {
	atomic.StoreUint32(&l.seconds, uint32(seconds))
}
