package limit

import (
	"context"
	"errors"
	"github.com/gotid/god/lib/store/redis"
	"strconv"
	"time"
)

// 为了兼容阿里云的 redis，我们无法使用 `local key = KEYS[1]` 来重用 key
const periodScript = `local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call("INCRBY", KEYS[1], 1)
if current == 1 then
    redis.call("expire", KEYS[1], window)
end
if current < limit then
    return 1
elseif current == limit then
    return 2
else
    return 0
end`

const (
	// Unknown 意为未初始化状态
	Unknown = iota
	// Allowed 意为允许请求状态
	Allowed
	// HitQuota 意为该请求达到时间段内的限额。
	HitQuota
	// OverQuota 意为超过限额。
	OverQuota

	internalOverQuota = 0
	internalAllowed   = 1
	internalHitQuota  = 2
)

// ErrUnknownCode 是一个表示位置状态的代码
var ErrUnknownCode = errors.New("未知状态码")

type (
	// PeriodLimit 用于在一段时间内限制请求。
	PeriodLimit struct {
		period     int          // 一段时间
		quota      int          // 时间内限额
		limitStore *redis.Redis // 后端存储
		keyPrefix  string       // 存储键前缀
		align      bool         // 是否开启时区对齐模式
	}

	// PeriodOption 自定义 PeriodLimit 的选项。
	PeriodOption func(pl *PeriodLimit)
)

// NewPeriodLimit 返回一个用于限速的 PeriodLimit。
// period 为一段以秒为单位的时间，quota 时间段内的请求限额数。
func NewPeriodLimit(period, quota int, limitStore *redis.Redis, keyPrefix string,
	opts ...PeriodOption) *PeriodLimit {

	limiter := &PeriodLimit{
		period:     period,
		quota:      quota,
		limitStore: limitStore,
		keyPrefix:  keyPrefix,
	}

	for _, opt := range opts {
		opt(limiter)
	}

	return limiter
}

// Take 获取给定 key 的持久化状态。
func (pl *PeriodLimit) Take(key string) (int, error) {
	return pl.TakeCtx(context.Background(), key)
}

// TakeCtx 获取给定 key 的持久化状态。
func (pl *PeriodLimit) TakeCtx(ctx context.Context, key string) (int, error) {
	resp, err := pl.limitStore.EvalCtx(ctx, periodScript, []string{pl.keyPrefix + key}, []string{
		strconv.Itoa(pl.quota),
		strconv.Itoa(pl.calcExpireSeconds()),
	})
	if err != nil {
		return Unknown, err
	}

	code, ok := resp.(int64)
	if !ok {
		return Unknown, ErrUnknownCode
	}

	switch code {
	case internalOverQuota:
		return OverQuota, nil
	case internalAllowed:
		return Allowed, nil
	case internalHitQuota:
		return HitQuota, nil
	default:
		return Unknown, ErrUnknownCode
	}
}

// 计算过期秒数。
func (pl *PeriodLimit) calcExpireSeconds() int {
	// 开启时区对齐模式
	if pl.align {
		now := time.Now()
		_, offset := now.Zone()
		unix := now.Unix() + int64(offset)
		return pl.period - int(unix%int64(pl.period))
	}

	// 非对齐模式
	return pl.period
}

// Align 自定义 PeriodLimit 的对齐模式。
// 如，我们想限制用户每天发送 5 条短信验证码，我们需要与当地时区保持一致。
func Align() PeriodOption {
	return func(pl *PeriodLimit) {
		pl.align = true
	}
}
