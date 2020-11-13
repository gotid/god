package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/mathx"
	"git.zc0901.com/go/god/lib/stat"
	"git.zc0901.com/go/god/lib/store/redis"
	"git.zc0901.com/go/god/lib/syncx"
	"math/rand"
	"sync"
	"time"
)

const (
	expiresDeviation    = 0.05 // 过期偏差
	notFoundPlaceholder = "*"  // 空记录占位符，防止缓存穿透
)

var errPlaceholder = errors.New("placeholder")

type node struct {
	redis           *redis.Redis
	barrier         syncx.SharedCalls
	expires         time.Duration
	notFoundExpires time.Duration
	unstableExpires mathx.Unstable
	stat            *Stat
	rnd             *rand.Rand
	lock            *sync.Mutex
	errNotFound     error
}

func NewCacheNode(r *redis.Redis, barrier syncx.SharedCalls, stat *Stat, errNotFound error, opts ...Option) Cache {
	o := newOptions(opts...)
	return node{
		redis:           r,
		barrier:         barrier,
		expires:         o.Expires,
		notFoundExpires: o.NotFoundExpires,
		unstableExpires: mathx.NewUnstable(expiresDeviation),
		stat:            stat,
		rnd:             rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:            new(sync.Mutex),
		errNotFound:     errNotFound,
	}
}

func (n node) Del(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if _, err := n.redis.Del(keys...); err != nil {
		logx.Errorf("删除缓存失败，keys: %q, 错误: %v", formatKeys(keys), err)
		n.asyncRetryDelCache(keys...)
	}

	return nil
}

func (n node) Get(key string, dest interface{}) error {
	if err := n.doGet(key, dest); err == errPlaceholder {
		return n.errNotFound
	} else {
		return err
	}
}

func (n node) Set(key string, value interface{}) error {
	return n.SetEx(key, value, n.aroundDuration(n.expires))
}

func (n node) SetEx(key string, value interface{}, expires time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return n.redis.SetEx(key, string(data), int(expires.Seconds()))
}

// Take 拿key对应的dest缓存，拿不到缓存就查库并缓存
func (n node) Take(dest interface{}, key string, queryFn func(interface{}) error) error {
	return n.doTake(dest, key, queryFn, func(value interface{}) error {
		return n.Set(key, value)
	})
}

// Take 读不到就写并设置有效期，然后返回
func (n node) TakeEx(dest interface{}, key string, queryFn func(interface{}, time.Duration) error) error {
	expires := n.aroundDuration(n.expires)
	return n.doTake(dest, key, func(value interface{}) error {
		// 读库
		return queryFn(value, expires)
	}, func(newVal interface{}) error {
		// 设置缓存有效期
		return n.SetEx(key, newVal, expires)
	})
}

func (n node) String() string {
	return n.redis.Addr
}

func (n node) asyncRetryDelCache(keys ...string) {
	AddCleanTask(func() error {
		_, err := n.redis.Del(keys...)
		return err
	}, keys...)
}

func (n node) doGet(key string, dest interface{}) error {
	n.stat.IncrTotal()
	result, err := n.redis.Get(key)
	if err != nil {
		n.stat.IncrMiss()
		return err
	}

	if len(result) == 0 {
		n.stat.IncrMiss()
		return n.errNotFound
	}

	n.stat.IncrHit()
	if result == notFoundPlaceholder {
		return errPlaceholder
	}

	return n.processCache(key, result, dest)
}

func (n node) doTake(dest interface{}, key string, queryFn func(newVal interface{}) error, cacheValFn func(newVal interface{}) error) error {
	// 防缓存击穿 barrier -> SharedCalls
	result, hit, err := n.barrier.Do(key, func() (interface{}, error) {
		if err := n.doGet(key, dest); err != nil {
			if err == errPlaceholder {
				return nil, n.errNotFound
			} else if err != n.errNotFound {
				// 直接返回错误而不是继续查库，以防高并发拖垮数据库
				return nil, err
			}

			// 查库
			if err := queryFn(dest); err == n.errNotFound {
				// 防缓存穿透
				if err = n.setWithNotFound(key); err != nil {
					logx.Error(err)
				}

				return nil, n.errNotFound
			} else if err != nil {
				n.stat.IncrDbFails()
				return nil, err
			}

			// 缓存数据库新查询值
			if err = cacheValFn(dest); err != nil {
				logx.Error(err)
			}
		}

		return json.Marshal(dest)
	})
	if err != nil {
		return err
	}
	if !hit {
		return nil
	}

	// 从之前查询的缓存中直接获取结果
	n.stat.IncrTotal()
	n.stat.IncrHit()

	return json.Unmarshal(result.([]byte), dest)
}

func (n node) processCache(key string, result string, dest interface{}) error {
	err := json.Unmarshal([]byte(result), dest)
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf("解封缓存失败，缓存节点：%s，键：%s，值：%s，错误：%v", n.redis.Addr, key, result, err)
	logx.Error(msg)
	stat.Report(msg)
	if _, err = n.redis.Del(key); err != nil {
		logx.Errorf("删除无效缓存，节点：%s，键：%s，值：%s，错误：%v", n.redis.Addr, key, result, err)
	}

	// 返回 errNotFound 以通过 queryFn 重新加载缓存值
	return n.errNotFound
}

// 防缓存雪崩：基于指定时间生成一个随机临近值，以防N多缓存同时过期，瞬间冲击数据库压力
func (n node) aroundDuration(expires time.Duration) time.Duration {
	return n.unstableExpires.AroundDuration(expires)
}

// 防缓存穿透：没找到的记录，照样缓存并设置短暂过期时间，减缓数据库压力
func (n node) setWithNotFound(key string) error {
	return n.redis.SetEx(key, notFoundPlaceholder, int(n.aroundDuration(n.notFoundExpires).Seconds()))
}
