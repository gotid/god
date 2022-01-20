package load

import (
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"git.zc0901.com/go/god/lib/collection"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/lib/syncx"
	"git.zc0901.com/go/god/lib/timex"
)

const (
	defaultBuckets = 50
	defaultWindow  = 5 * time.Second

	// 1000m 表示法，900m 差不多相当于80%
	defaultCpuThreshold = 900

	// 默认最小响应时间
	defaultMinRt = float64(time.Second / time.Millisecond)

	flyingBeta      = 0.9         // 用于动态计算的超参
	coolOffDuration = time.Second // 冷却时长
)

var (
	ErrServiceOverloaded = errors.New("服务超载")

	// 是否启用泄流阀，默认启用。
	enabled = syncx.ForAtomicBool(true)
	// 是否启用泄流阀日志统计，默认启用。
	logEnabled = syncx.ForAtomicBool(true)
	// 超载检测函数（判断CPU用量是否超过预置的阈值）。
	systemOverloadChecker = func(cpuThreshold int64) bool {
		return stat.CpuUsage() >= cpuThreshold
	}
)

type (
	// Promise 接口用于 Shedder.Allow 告知调用方请求是否成功。
	Promise interface {
		Pass() // 告知调用者调用成功。
		Fail() // 告知调用者调用失败。
	}

	// Shedder 表示一个负载泄流阀。
	Shedder interface {
		Allow() (Promise, error) // 允许返回 Promise，不允许返 ErrServiceOverloaded。
	}

	// ShedderOption 自定义 Shedder 的函数。
	ShedderOption func(opts *shedderOptions)

	// Shedder 的自定义项。
	shedderOptions struct {
		window       time.Duration
		buckets      int
		cpuThreshold int64
	}

	// 自适应泄流阀。
	adaptiveShedder struct {
		cpuThreshold    int64
		windows         int64 // 每秒的buckets
		flying          int64 // 飞行架次
		avgFlying       float64
		avgFlyingLock   syncx.SpinLock
		dropTime        *syncx.AtomicDuration
		droppedRecently *syncx.AtomicBool
		passCounter     *collection.RollingWindow // 请求通过计数器
		rtCounter       *collection.RollingWindow // 响应时间计数器
	}
)

// NewAdaptiveShedder 返回自适应泄流阀。
// opts 用于自定义 Shedder。
func NewAdaptiveShedder(opts ...ShedderOption) Shedder {
	if !enabled.True() {
		return newNopShedder()
	}

	options := shedderOptions{
		window:       defaultWindow,
		buckets:      defaultBuckets,
		cpuThreshold: defaultCpuThreshold,
	}
	for _, opt := range opts {
		opt(&options)
	}
	bucketDuration := options.window / time.Duration(options.buckets)
	return &adaptiveShedder{
		cpuThreshold:    options.cpuThreshold,
		windows:         int64(time.Second / bucketDuration),
		dropTime:        syncx.NewAtomicDuration(),
		droppedRecently: syncx.NewAtomicBool(),
		passCounter:     collection.NewRollingWindow(options.buckets, bucketDuration, collection.IgnoreCurrentBucket()),
		rtCounter:       collection.NewRollingWindow(options.buckets, bucketDuration, collection.IgnoreCurrentBucket()),
	}
}

// WithBuckets 自定义泄流阀的桶数。
func WithBuckets(buckets int) ShedderOption {
	return func(opts *shedderOptions) {
		opts.buckets = buckets
	}
}

// WithWindow 自定义泄流阀的窗口时长。
func WithWindow(window time.Duration) ShedderOption {
	return func(opts *shedderOptions) {
		opts.window = window
	}
}

// WithCpuThreshold 自定义泄流阀的 cpu 阈值。
func WithCpuThreshold(cpuThreshold int64) ShedderOption {
	return func(opts *shedderOptions) {
		opts.cpuThreshold = cpuThreshold
	}
}

// Disable 禁用负载泄流阀。
func Disable() {
	enabled.Set(false)
}

// DisableLog 禁用泄流阀日志统计。
func DisableLog() {
	logEnabled.Set(false)
}

// Allow 判断是否接受请求并进行相关处理。
func (s *adaptiveShedder) Allow() (Promise, error) {
	if s.shouldDrop() {
		s.dropTime.Set(timex.Now())
		s.droppedRecently.Set(true)

		return nil, ErrServiceOverloaded
	}

	s.addFlying(1)

	return &promise{
		start:   timex.Now(),
		shedder: s,
	}, nil
}

// shouldDrop 判断是否删除该请求，若删除则记录日志
func (s *adaptiveShedder) shouldDrop() bool {
	if s.systemOverloaded() || s.stillHot() {
		if s.highThru() {
			flying := atomic.LoadInt64(&s.flying)
			s.avgFlyingLock.Lock()
			avgFlying := s.avgFlying
			s.avgFlyingLock.Unlock()
			msg := fmt.Sprintf("丢弃请求，CPU: %d, maxPass: %d, minRt: %.2f, hot: %t, flying: %d, avgFlying: %.2f",
				stat.CpuUsage(), s.maxPass(), s.minRt(), s.stillHot(), flying, avgFlying)
			logx.Error(msg)
			stat.Report(msg)
			return true
		}
	}

	return false
}

func (s *adaptiveShedder) systemOverloaded() bool {
	return systemOverloadChecker(s.cpuThreshold)
}

func (s *adaptiveShedder) stillHot() bool {
	if !s.droppedRecently.True() {
		return false
	}

	dropTime := s.dropTime.Load()
	if dropTime == 0 {
		return false
	}

	hot := timex.Since(dropTime) < coolOffDuration
	if !hot {
		s.droppedRecently.Set(false)
	}

	return hot
}

func (s *adaptiveShedder) highThru() bool {
	s.avgFlyingLock.Lock()
	avgFlying := s.avgFlying
	s.avgFlyingLock.Unlock()
	maxFlight := s.maxFlight()
	return int64(avgFlying) > maxFlight && atomic.LoadInt64(&s.flying) > maxFlight
}

func (s *adaptiveShedder) maxFlight() int64 {
	// windows = buckets per second
	// maxQPS = maxPASS * windows
	// minRT = 最小平均响应时间(毫秒)
	// maxQPS = minRT / 每秒的毫秒数
	return int64(math.Max(1, float64(s.maxPass()*s.windows)*(s.minRt()/1e3)))
}

// maxPass 最大请求数
func (s *adaptiveShedder) maxPass() int64 {
	var result float64 = 1

	s.passCounter.Reduce(func(b *collection.Bucket) {
		if b.Accepts > result {
			result = b.Accepts
		}
	})

	return int64(result)
}

// minRt 最小平均响应时间(毫秒)
func (s *adaptiveShedder) minRt() float64 {
	result := defaultMinRt

	s.rtCounter.Reduce(func(b *collection.Bucket) {
		if b.Requests <= 0 {
			return
		}

		avg := math.Round(b.Accepts / float64(b.Requests))
		if avg < result {
			result = avg
		}
	})

	return result
}

func (s *adaptiveShedder) addFlying(delta int64) {
	flying := atomic.AddInt64(&s.flying, delta)
	// 当请求完成，更新 avgFlying
	// 该策略让 avgFlying 和 flying 稍有延迟，更加平滑：
	// 1. 当飞行请求快速增加时，avgFlying 增加较慢，可接受更多请求，
	// 2. 当飞行请求大量被删时，avgFlying 慢慢地删，让无效请求更少，
	// 如此，让服务可以尽可能多的接收更多请求。

	if delta < 0 {
		s.avgFlyingLock.Lock()
		s.avgFlying = s.avgFlying*flyingBeta + float64(flying)*(1-flyingBeta)
		s.avgFlyingLock.Unlock()
	}
}

// Promise 接口的实现结构体。
type promise struct {
	start   time.Duration
	shedder *adaptiveShedder
}

func (p *promise) Pass() {
	rt := float64(timex.Since(p.start)) / float64(time.Millisecond)
	p.shedder.addFlying(-1)
	p.shedder.rtCounter.Add(math.Ceil(rt))
	p.shedder.passCounter.Add(1)
}

func (p *promise) Fail() {
	p.shedder.addFlying(-1)
}
