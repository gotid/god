package load

import (
	"errors"
	"fmt"
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
	"math"
	"sync/atomic"
	"time"
)

const (
	// 默认窗口时长
	defaultWindow = 5 * time.Second
	// 默认每个窗口的桶数
	defaultBuckets = 50
	// 默认CPU阈值，使用 1000m 注解，900m 类似 80%
	defaultCpuThreshold = 900
	// 默认最小响应时间
	defaultMinRt = float64(time.Second / time.Millisecond)
	// 用于计算动态请求的移动平均超参数beta
	flyingBeta     = 0.9
	coolOfDuration = time.Second
)

var (
	// ErrServiceOverloaded 由 Shedder.Allow 在服务发生超载时返回
	ErrServiceOverloaded = errors.New("服务已超载")

	//默认启用自动降载
	enabled = syncx.ForAtomicBool(true)
	//默认启用自动降载的统计日志
	logEnabled = syncx.ForAtomicBool(true)
	//检测当前 cpu 阈值是否过载
	systemOverloadChecker = func(cpuThreshold int64) bool {
		return stat.CpuUsage() >= cpuThreshold
	}
)

type (
	// Promise 是 Shedder 返回的接口。
	// 让调用者判断请求是否处理成功。
	Promise interface {
		// Pass 告诉调用者调用成功。
		Pass()
		// Fail 告诉调用者调用失败。
		Fail()
	}

	// Shedder 是包装 Allow 方法的接口。
	Shedder interface {
		// Allow 如果允许，则返回 Promise，否则返回 ErrorServiceOverloaded。
		Allow() (Promise, error)
	}

	// ShedderOption 自定义 Shedder 的方法。
	ShedderOption func(opts *shedderOptions)

	shedderOptions struct {
		window       time.Duration // 泄流器窗口时长
		buckets      int
		cpuThreshold int64
	}

	adaptiveShedder struct {
		cpuThreshold    int64
		windows         int64
		flying          int64
		avgFlying       float64
		avgFlyingLock   syncx.SpinLock
		overloadTime    *syncx.AtomicDuration
		droppedRecently *syncx.AtomicBool
		passCounter     *collection.RollingWindow
		rtCounter       *collection.RollingWindow
	}
)

// Disable 禁用自动降载。
func Disable() {
	enabled.Set(false)
}

// DisableLog 禁用自动降载器的统计日志。
func DisableLog() {
	logEnabled.Set(false)
}

// NewAdaptiveShedder 返回一个自适应的 CPU 自动降载器。
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
		overloadTime:    syncx.NewAtomicDuration(),
		droppedRecently: syncx.NewAtomicBool(),
		passCounter:     collection.NewRollingWindow(options.buckets, bucketDuration, collection.IgnoreCurrentBucket()),
		rtCounter:       collection.NewRollingWindow(options.buckets, bucketDuration, collection.IgnoreCurrentBucket()),
	}
}

// Allow 实现 Shedder.Allow 方法。
func (as *adaptiveShedder) Allow() (Promise, error) {
	if as.shouldDrop() {
		as.droppedRecently.Set(true)

		return nil, ErrServiceOverloaded
	}

	as.addFlying(1)

	return &promise{
		start:   timex.Now(),
		shedder: as,
	}, nil
}

func (as *adaptiveShedder) shouldDrop() bool {
	if as.systemOverloaded() || as.stillHot() {
		if as.highThru() {
			flying := atomic.LoadInt64(&as.flying)
			as.avgFlyingLock.Lock()
			avgFlying := as.avgFlying
			as.avgFlyingLock.Unlock()

			msg := fmt.Sprintf("dropreq, cpu: %d, maxPass: %d, minRt: %.2f, hot: %t, flying: %d, avgFlying: %.2f",
				stat.CpuUsage(), as.maxPass(), as.minRt(), as.stillHot(), flying, avgFlying)
			logx.Error(msg)
			stat.Report(msg)
			return true
		}
	}

	return false
}

func (as *adaptiveShedder) addFlying(delta int64) {
	flying := atomic.AddInt64(&as.flying, delta)
	// 请求完成后更新 avgFlying
	// 该策略使得 avgFlying 相比 flying 有点滞后，但更流畅。
	// 当 flying 请求快速增加时，addFlying 增加较慢，接受的请求较多。
	// 当 flying 请求快速下降时，addFlying 下降较慢，接受的请求较少。
	// 该策略使得服务可以接受尽可能多的请求。
	if delta < 0 {
		as.avgFlyingLock.Lock()
		as.avgFlying = as.avgFlying*flyingBeta + float64(flying)*(1-flyingBeta)
		as.avgFlyingLock.Unlock()
	}
}

func (as *adaptiveShedder) systemOverloaded() bool {
	if !systemOverloadChecker(as.cpuThreshold) {
		return false
	}

	as.overloadTime.Set(timex.Now())
	return true
}

func (as *adaptiveShedder) stillHot() bool {
	if !as.droppedRecently.True() {
		return false
	}

	overloadTime := as.overloadTime.Load()
	if overloadTime == 0 {
		return false
	}

	hot := timex.Since(overloadTime) < coolOfDuration
	if !hot {
		as.droppedRecently.Set(false)
	}

	return hot
}

func (as *adaptiveShedder) highThru() bool {
	as.avgFlyingLock.Lock()
	avgFlying := as.avgFlying
	as.avgFlyingLock.Unlock()

	maxFlight := as.maxFlight()
	return int64(avgFlying) > maxFlight && atomic.LoadInt64(&as.flying) > maxFlight
}

func (as *adaptiveShedder) maxFlight() int64 {
	// windows = 每秒的桶数
	// maxQPS = maxPASS * windows
	// minRT = 毫秒单位的最小平均响应时间
	// maxQPS * minRT / 每秒的毫秒数
	return int64(math.Max(1, float64(as.maxPass()*as.windows)*(as.minRt()/1e3)))
}

func (as *adaptiveShedder) maxPass() int64 {
	var result float64 = 1

	as.passCounter.Reduce(func(b *collection.Bucket) {
		if b.Sum > result {
			result = b.Sum
		}
	})

	return int64(result)
}

func (as *adaptiveShedder) minRt() float64 {
	result := defaultMinRt

	as.rtCounter.Reduce(func(b *collection.Bucket) {
		if b.Count <= 0 {
			return
		}

		avg := math.Round(b.Sum / float64(b.Count))
		if avg < result {
			result = avg
		}
	})

	return result
}

// WithWindow 自定义泄流器的窗口时长。
func WithWindow(window time.Duration) ShedderOption {
	return func(opts *shedderOptions) {
		opts.window = window
	}
}

// WithBuckets 自定义泄流器每个窗口的桶数。
func WithBuckets(buckets int) ShedderOption {
	return func(opts *shedderOptions) {
		opts.buckets = buckets
	}
}

// WithCpuThreshold 自定义自动降载器的 cpu 阈值。
func WithCpuThreshold(threshold int64) ShedderOption {
	return func(opts *shedderOptions) {
		opts.cpuThreshold = threshold
	}
}

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
