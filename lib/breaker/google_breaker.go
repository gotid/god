package breaker

import (
	"math"
	"time"

	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/mathx"
)

const (
	window  = 10 * time.Second // 一个滚动窗时长，默认10秒
	buckets = 40               // 一个滚动窗允许通过的桶数，默认40个

	K          = 1.5 // 请求接受比例，越大接受度则越高，越小自适应节流则越积极
	protection = 5   // 自我保护的请求数
)

type (
	// googleBreaker 是谷歌处理 overload 过载问题的节流阀实现。
	// see Client-Side Throttling section in https://landing.google.com/sre/sre-book/chapters/handling-overload/
	googleBreaker struct {
		k     float64                   // 请求接受比例
		stat  *collection.RollingWindow // 统计窗口计数器（采用滚窗算法）
		proba *mathx.Proba
	}

	googlePromise struct {
		b *googleBreaker
	}
)

func newGoogleBreaker() *googleBreaker {
	bucketDuration := time.Duration(int64(window) / int64(buckets)) // 单桶时长，默认250毫秒
	statWindow := collection.NewRollingWindow(buckets, bucketDuration)
	return &googleBreaker{
		k:     K,
		stat:  statWindow,
		proba: mathx.NewProba(),
	}
}

// 先看断路器是否接受，如果接受则发返回 googlePromise 等待处理
func (b *googleBreaker) allow() (internalPromise, error) {
	if err := b.accept(); err != nil {
		return nil, err
	}

	// 接受成功，则返回googlePromise，由其标记结果
	return googlePromise{b: b}, nil
}

func (b *googleBreaker) doReq(req Request, fallback Fallback, acceptable Acceptable) error {
	// 首先，试探断路器是否接受请求
	if err := b.accept(); err != nil {
		// 尝试采用应急方案
		if fallback != nil {
			return fallback(err)
		} else {
			return err
		}
	}

	// 最后，如果有错则标记为失败，并抛出异常
	defer func() {
		if e := recover(); e != nil {
			b.markFailure()
			panic(e)
		}
	}()

	// 然后，执行请求，根据返回错误的可接受度标记结果
	err := req()
	if acceptable(err) {
		b.markSuccess()
	} else {
		b.markFailure()
	}

	// 不论错误是否可接受，都要返回
	return err
}

// accept 根据客户端请求拒绝率返回错误
func (b *googleBreaker) accept() error {
	requests, accepts := b.history()

	// 计算客户端请求拒绝率
	// https://landing.google.com/sre/sre-book/chapters/handling-overload/#eq2101
	// 常量 K 为1.5意味着，请求150次只成功100次，则
	weightedAccepts := b.k * float64(accepts)
	droppedRequests := float64(requests-protection) - weightedAccepts
	dropRatio := math.Max(0, droppedRequests/float64(requests+1))

	// 无需拒绝
	if dropRatio <= 0 {
		return nil
	}

	// 并非每次阻断，而是随机拦截，以此给后端重生的机会
	if b.proba.TrueOnProba(dropRatio) {
		return ErrServiceUnavailable
	}

	return nil
}

// 历史总数（请求多少次，同意多少次）
func (b *googleBreaker) history() (requests int64, accepts int64) {
	b.stat.Reduce(func(b *collection.Bucket) {
		requests += b.Requests
		accepts += int64(b.Accepts)
	})
	return
}

func (b *googleBreaker) markSuccess() {
	b.stat.Add(1)
}

func (b *googleBreaker) markFailure() {
	b.stat.Add(0)
}

func (p googlePromise) Accept() {
	p.b.markSuccess()
}

func (p googlePromise) Reject() {
	p.b.markFailure()
}
