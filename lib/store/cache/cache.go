package cache

import (
	"context"
	"github.com/gotid/god/lib/hash"
	"github.com/gotid/god/lib/syncx"
	"log"
	"time"
)

// Cache 定义了缓存层实现删除、获取、设置、回源的接口。
type Cache interface {
	// Del 删除给定键名的缓存值。
	Del(keys ...string) error
	// DelCtx 删除给定键名的缓存值。
	DelCtx(ctx context.Context, keys ...string) error
	// Get 获取给定 key 的缓存并填充至 val。
	Get(key string, val any) error
	// GetCtx 获取给定 key 的缓存并填充至 val。
	GetCtx(ctx context.Context, key string, val any) error
	// IsNotFound 判断给定错误是否为预定义的未找到错误。
	IsNotFound(err error) bool
	// Set 设置键值对缓存，并将其存活时间设置为 n.expire。
	Set(key string, val any) error
	// SetCtx 设置键值对缓存，并将其存活时间设置为 n.expire。
	SetCtx(ctx context.Context, key string, val any) error
	// SetWithExpire 设置给定的键值对及过期时长。
	SetWithExpire(key string, val any, expire time.Duration) error
	// SetWithExpireCtx 设置给定的键值对及过期时长。
	SetWithExpireCtx(ctx context.Context, key string, val any, expire time.Duration) error
	// Take 首先从缓存中获取结果，如果未找到则从DB查询并设置过期时长，然后返回结果。
	Take(val any, key string, query func(val any) error) error
	// TakeCtx 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
	TakeCtx(ctx context.Context, val any, key string, query func(val any) error) error
	// TakeWithExpire 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
	TakeWithExpire(val any, key string, query func(val any, expire time.Duration) error) error
	// TakeWithExpireCtx 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
	TakeWithExpireCtx(ctx context.Context, val any, key string, query func(val any, expire time.Duration) error) error
}

// New 返回一个缓存 Cache。
func New(c Config, barrier syncx.SingleFlight, st *Stat, errNotFound error, opts ...Option) Cache {
	if len(c) == 0 || TotalWeights(c) <= 0 {
		log.Fatal("未配置缓存节点")
	}

	if len(c) == 1 {
		return NewNode(c[0].NewRedis(), barrier, st, errNotFound, opts...)
	}

	dispatcher := hash.NewConsistentHash()
	for _, cfg := range c {
		n := NewNode(cfg.NewRedis(), barrier, st, errNotFound, opts...)
		dispatcher.AddWithWeight(n, cfg.Weight)
	}

	return cluster{
		dispatcher:  dispatcher,
		errNotFound: errNotFound,
	}
}
