package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/jsonx"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/syncx"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	// 设置过期公差值以防大量缓存项同时过期造成雪崩
	// 设为不固定的缓存描述为区间 [0.95, 1.05] * 秒
	expireDeviation     = 0.05
	notFoundPlaceholder = "*"
)

// 代表关联的键还没有值
var errPlaceholder = errors.New("缓存占位符")

type node struct {
	rds            *redis.Redis
	expire         time.Duration
	notFoundExpire time.Duration
	barrier        syncx.SingleFlight
	r              *rand.Rand
	lock           *sync.Mutex
	unstableExpire mathx.Unstable
	stat           *Stat
	errNotFound    error
}

// NewNode 返回一个缓存节点 node。
func NewNode(rds *redis.Redis, barrier syncx.SingleFlight, st *Stat, errNotFound error, opts ...Option) Cache {
	o := newOptions(opts...)
	return node{
		rds:            rds,
		expire:         o.Expire,
		notFoundExpire: o.NotFoundExpire,
		barrier:        barrier,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           st,
		errNotFound:    errNotFound,
	}
}

// Del 删除给定 keys 的缓存。
func (n node) Del(keys ...string) error {
	return n.DelCtx(context.Background(), keys...)
}

// DelCtx 删除给定 keys 的缓存。
func (n node) DelCtx(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	logger := logx.WithContext(ctx)
	if len(keys) > 1 && n.rds.Type == redis.ClusterType {
		for _, key := range keys {
			if _, err := n.rds.DelCtx(ctx, key); err != nil {
				logger.Errorf("未能清除缓存键：%q，错误：%v", key, err)
				n.asyncRetryDelCache(key)
			}
		}
	} else if _, err := n.rds.DelCtx(ctx, keys...); err != nil {
		logger.Errorf("未能清除缓存键：%q，错误：%v", formatKeys(keys), err)
		n.asyncRetryDelCache(keys...)
	}

	return nil
}

// Get 获取给定 key 的缓存并填充至 val。
func (n node) Get(key string, val interface{}) error {
	return n.GetCtx(context.Background(), key, val)
}

// GetCtx 获取给定 key 的缓存并填充至 val。
func (n node) GetCtx(ctx context.Context, key string, val interface{}) error {
	err := n.doGetCache(ctx, key, val)
	if err == errPlaceholder {
		return n.errNotFound
	}

	return err
}

// IsNotFound 判断给定错误是否为预定义的未找到错误。
func (n node) IsNotFound(err error) bool {
	return errors.Is(err, n.errNotFound)
}

// Set 设置键值对缓存，并将其存活时间设置为 n.expire。
func (n node) Set(key string, val interface{}) error {
	return n.SetCtx(context.Background(), key, val)
}

// SetCtx 设置键值对缓存，并将其存活时间设置为 n.expire。
func (n node) SetCtx(ctx context.Context, key string, val interface{}) error {
	return n.SetWithExpireCtx(ctx, key, val, n.aroundDuration(n.expire))
}

// SetWithExpire 设置给定的键值对及过期时长。
func (n node) SetWithExpire(key string, val interface{}, expire time.Duration) error {
	return n.SetWithExpireCtx(context.Background(), key, val, expire)
}

// SetWithExpireCtx 设置给定的键值对及过期时长。
func (n node) SetWithExpireCtx(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	data, err := jsonx.Marshal(val)
	if err != nil {
		return err
	}

	return n.rds.SetexCtx(ctx, key, string(data), int(math.Ceil(expire.Seconds())))
}

// String 返回一个节点的字符串表示形式。
func (n node) String() string {
	return n.rds.Addr
}

// Take 首先从缓存中获取结果，如果未找到则从DB查询并设置过期时长，然后返回结果。
func (n node) Take(val interface{}, key string, query func(val interface{}) error) error {
	return n.TakeCtx(context.Background(), val, key, query)
}

// TakeCtx 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
func (n node) TakeCtx(ctx context.Context, val interface{}, key string, query func(val interface{}) error) error {
	return n.doTake(ctx, val, key, query, func(v interface{}) error {
		return n.SetCtx(ctx, key, val)
	})
}

// TakeWithExpire 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
func (n node) TakeWithExpire(val interface{}, key string, query func(val interface{}, expire time.Duration) error) error {
	return n.TakeWithExpireCtx(context.Background(), val, key, query)
}

// TakeWithExpireCtx 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
func (n node) TakeWithExpireCtx(ctx context.Context, val interface{}, key string, query func(val interface{}, expire time.Duration) error) error {
	expire := n.aroundDuration(n.expire)
	return n.doTake(ctx, val, key, func(val interface{}) error {
		return query(val, expire)
	}, func(val interface{}) error {
		return n.SetWithExpireCtx(ctx, key, val, expire)
	})
}

func (n node) asyncRetryDelCache(keys ...string) {
	AddCleanTask(func() error {
		_, err := n.rds.Del(keys...)
		return err
	}, keys...)
}

func (n node) doGetCache(ctx context.Context, key string, val interface{}) error {
	n.stat.IncrTotal()
	data, err := n.rds.GetCtx(ctx, key)
	if err != nil {
		n.stat.IncrMiss()
		return err
	}

	if len(data) == 0 {
		n.stat.IncrMiss()
		return n.errNotFound
	}

	n.stat.IncrHit()
	if data == notFoundPlaceholder {
		return errPlaceholder
	}

	return n.processCache(ctx, key, data, val)
}

func (n node) doTake(ctx context.Context, val interface{}, key string,
	query func(val interface{}) error,
	cacheVal func(val interface{}) error) error {
	logger := logx.WithContext(ctx)
	data, fresh, err := n.barrier.DoEx(key, func() (interface{}, error) {
		if err := n.doGetCache(ctx, key, val); err != nil {
			if err == errPlaceholder {
				return nil, n.errNotFound
			}

			if err = query(val); err == n.errNotFound {
				if err = n.setCacheWithNotFound(ctx, key); err != nil {
					logger.Error(err)
				}

				return nil, n.errNotFound
			} else if err != nil {
				n.stat.IncrDbFails()
				return nil, err
			}

			if err = cacheVal(val); err != nil {
				logger.Error(err)
			}
		}

		return jsonx.Marshal(val)
	})
	if err != nil {
		return err
	}
	if fresh {
		return nil
	}

	// 从上一次查询获取结果。
	// 为何不在函数一开始就 IncrTotal？
	// 因为返回了一个共享的错误，而且我们不想对其计数。
	// 例如，数据库挂了，query查询将失败，我们只想记录一次数据库失败。
	n.stat.IncrTotal()
	n.stat.IncrHit()

	return jsonx.Unmarshal(data.([]byte), val)
}

func (n node) processCache(ctx context.Context, key, data string, val interface{}) error {
	err := jsonx.Unmarshal([]byte(data), val)
	if err == nil {
		return nil
	}

	report := fmt.Sprintf("解编排缓存失败，节点：%s，键：%s，值：%s，错误：%v",
		n.rds.Addr, key, data, err)
	logger := logx.WithContext(ctx)
	logger.Error(report)
	stat.Report(report)
	if _, e := n.rds.DelCtx(ctx, key); e != nil {
		logger.Errorf("删除无效缓存失败，节点：%s，键：%s，值：%s，错误：%v",
			n.rds.Addr, key, data, e)
	}

	// 返回 errNotFound 可根据给定的 queryFn 重载数据
	return n.errNotFound
}

func (n node) aroundDuration(expire time.Duration) time.Duration {
	return n.unstableExpire.AroundDuration(expire)
}

func (n node) setCacheWithNotFound(ctx context.Context, key string) error {
	seconds := int(math.Ceil(n.aroundDuration(n.notFoundExpire).Seconds()))
	return n.rds.SetexCtx(ctx, key, notFoundPlaceholder, seconds)
}
