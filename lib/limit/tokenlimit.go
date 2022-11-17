package limit

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/logx"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/store/redis"
	xrate "golang.org/x/time/rate"
)

const (
	tokenFormat     = "{%s}.tokens"
	timestampFormat = "{%s}.ts"

	// 为了兼容阿里云的 redis，我们无法使用 `local key = KEYS[1]` 来重用 key
	// KEYS[1] as tokens_key
	// KEYS[2] as timestamp_key
	script = `local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = capacity/rate
local ttl = math.floor(fill_time*2)
local last_tokens = tonumber(redis.call("get", KEYS[1]))
if last_tokens == nil then
    last_tokens = capacity
end

local last_refreshed = tonumber(redis.call("get", KEYS[2]))
if last_refreshed == nil then
    last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local filled_tokens = math.min(capacity, last_tokens+(delta*rate))
local allowed = filled_tokens >= requested
local new_tokens = filled_tokens
if allowed then
    new_tokens = filled_tokens - requested
end

redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)

return allowed`

	pingInterval = 100 * time.Millisecond
)

// TokenLimiter 速率限制器：用于控制一秒钟内允许事件发生的频率。
type TokenLimiter struct {
	rate           int
	burst          int
	store          *redis.Redis
	tokenKey       string
	timestampKey   string
	rescueLock     sync.Mutex
	redisAlive     uint32
	monitorStarted bool
	rescueLimiter  *xrate.Limiter
}

// NewTokenLimiter 返回一个新的 TokenLimiter，它允许事件达到最高速率数， 并允许突发令牌数。
// rate 为最高速率数，burst 突发令牌数。
func NewTokenLimiter(rate, burst int, store *redis.Redis, key string) *TokenLimiter {
	tokenKey := fmt.Sprintf(tokenFormat, key)
	timestampKey := fmt.Sprintf(timestampFormat, key)

	return &TokenLimiter{
		rate:           rate,
		burst:          burst,
		store:          store,
		tokenKey:       tokenKey,
		timestampKey:   timestampKey,
		rescueLock:     sync.Mutex{},
		redisAlive:     1,
		monitorStarted: false,
		rescueLimiter:  xrate.NewLimiter(xrate.Every(time.Second/time.Duration(rate)), burst),
	}
}

// Allow 是 AllowN(time.Now(), 1) 的简写方式。
func (tl *TokenLimiter) Allow() bool {
	return tl.AllowN(time.Now(), 1)
}

// AllowCtx 是 AllowNCtx(ctx, time.Now(), 1) 的简写方式。
func (tl *TokenLimiter) AllowCtx(ctx context.Context) bool {
	return tl.AllowNCtx(ctx, time.Now(), 1)
}

// AllowN 判断此时是否会发生 n 个事件。
// 如果打算丢弃或跳过超速率事件，就使用此法；
// 否则，使用 Reserve 或 Wait 方法。
func (tl *TokenLimiter) AllowN(now time.Time, n int) bool {
	return tl.reserveN(context.Background(), now, n)
}

// AllowNCtx 判断此时是否会发生 n 个时间。
// 如果打算丢弃或跳过超速率事件，就使用此法；
// 否则，使用 Reserve 或 Wait 方法。
func (tl *TokenLimiter) AllowNCtx(ctx context.Context, now time.Time, n int) bool {
	return tl.reserveN(ctx, now, n)
}

func (tl *TokenLimiter) reserveN(ctx context.Context, now time.Time, n int) bool {
	if atomic.LoadUint32(&tl.redisAlive) == 0 {
		return tl.rescueLimiter.AllowN(now, n)
	}

	resp, err := tl.store.EvalCtx(ctx,
		script,
		[]string{
			tl.tokenKey,
			tl.timestampKey,
		},
		[]string{
			strconv.Itoa(tl.rate),
			strconv.Itoa(tl.burst),
			strconv.FormatInt(now.Unix(), 10),
			strconv.Itoa(n),
		},
	)

	// redis allowed == false
	// Lua boolean false -> r 响应 Nil
	if err == redis.Nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		logx.Errorf("无法使用速率限制器：%s", err)
		return false
	}

	if err != nil {
		logx.Errorf("无法使用速率限制器：%s，使用进程内限制器进行替补", err)
		tl.startMonitor()
		return tl.rescueLimiter.AllowN(now, n)
	}

	code, ok := resp.(int64)
	if !ok {
		logx.Errorf("无法评估 redis 脚本：%v，使用进程内限制器进行替补", resp)
		tl.startMonitor()
		return tl.rescueLimiter.AllowN(now, n)
	}

	// redis allowed = true
	// Lua boolean true -> r 响应整数 1
	return code == 1
}

func (tl *TokenLimiter) startMonitor() {
	tl.rescueLock.Lock()
	defer tl.rescueLock.Unlock()

	if tl.monitorStarted {
		return
	}

	tl.monitorStarted = true
	atomic.StoreUint32(&tl.redisAlive, 0)

	go tl.waitForRedis()
}

func (tl *TokenLimiter) waitForRedis() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		tl.rescueLock.Lock()
		tl.monitorStarted = false
		tl.rescueLock.Unlock()
	}()

	for range ticker.C {
		if tl.store.Ping() {
			atomic.StoreUint32(&tl.redisAlive, 1)
			return
		}
	}
}
