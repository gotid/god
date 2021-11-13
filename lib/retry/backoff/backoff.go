package backoff

import (
	"math/rand"
	"time"
)

// Func 定义计算重试时间的函数。
type Func func(attempt int) time.Duration

// LinearWithJitter 等待一段设定的时间，允许抖动（分数调整）。
func LinearWithJitter(waitBetween time.Duration, jitterFraction float64) Func {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func(attempt int) time.Duration {
		multiplier := jitterFraction * (r.Float64()*2 - 1)
		return time.Duration(float64(waitBetween) * (1 + multiplier))
	}
}

// Interval 在两次调用之间，等待固定时长。
func Interval(interval time.Duration) Func {
	return func(attempt int) time.Duration {
		return interval
	}
}

// Exponential 在两次调用之间，等待时长指数级递增。
func Exponential(scalar time.Duration) Func {
	return func(attempt int) time.Duration {
		return scalar * time.Duration((1<<attempt)>>1)
	}
}
