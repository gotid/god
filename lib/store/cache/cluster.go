package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/errorx"
	"github.com/gotid/god/lib/hash"
	"time"
)

// cluster 是一个支持一致性哈希分发的缓存集群
type cluster struct {
	dispatcher  *hash.ConsistentHash
	errNotFound error
}

// Del 删除给定 keys 的缓存。
func (c cluster) Del(keys ...string) error {
	return c.DelCtx(context.Background(), keys...)
}

// DelCtx 删除给定 keys 的缓存。
func (c cluster) DelCtx(ctx context.Context, keys ...string) error {
	switch len(keys) {
	case 0:
		return nil
	case 1:
		key := keys[0]
		n, ok := c.dispatcher.Get(key)
		if !ok {
			return c.errNotFound
		}

		return n.(Cache).DelCtx(ctx, key)
	default:
		var be errorx.BatchError
		nodes := make(map[any][]string)
		for _, key := range keys {
			n, ok := c.dispatcher.Get(key)
			if !ok {
				be.Add(fmt.Errorf("缓存键 %q 未找到", key))
				continue
			}

			nodes[n] = append(nodes[n], key)
		}
		for n, ks := range nodes {
			if err := n.(Cache).DelCtx(ctx, ks...); err != nil {
				be.Add(err)
			}
		}

		return be.Err()
	}
}

// Get 获取给定 key 的缓存并填充至 val。
func (c cluster) Get(key string, val any) error {
	return c.GetCtx(context.Background(), key, val)
}

// GetCtx 获取给定 key 的缓存并填充至 val。
func (c cluster) GetCtx(ctx context.Context, key string, val any) error {
	n, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return n.(Cache).GetCtx(ctx, key, val)
}

// IsNotFound 判断给定错误是否为预定义的未找到错误。
func (c cluster) IsNotFound(err error) bool {
	return errors.Is(err, c.errNotFound)
}

// Set 设置键值对缓存，并将其存活时间设置为 c.expire。
func (c cluster) Set(key string, val any) error {
	return c.SetCtx(context.Background(), key, val)
}

// SetCtx 设置键值对缓存，并将其存活时间设置为 c.expire。
func (c cluster) SetCtx(ctx context.Context, key string, val any) error {
	n, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return n.(Cache).SetCtx(ctx, key, val)
}

// SetWithExpire 设置给定的键值对及过期时长。
func (c cluster) SetWithExpire(key string, val any, expire time.Duration) error {
	return c.SetWithExpireCtx(context.Background(), key, val, expire)
}

// SetWithExpireCtx 设置给定的键值对及过期时长。
func (c cluster) SetWithExpireCtx(ctx context.Context, key string, val any, expire time.Duration) error {
	n, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return n.(Cache).SetWithExpireCtx(ctx, key, val, expire)
}

// Take 首先从缓存中获取结果，如果未找到则从DB查询并设置过期时长，然后返回结果。
func (c cluster) Take(val any, key string, query func(val any) error) error {
	return c.TakeCtx(context.Background(), val, key, query)
}

// TakeCtx 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
func (c cluster) TakeCtx(ctx context.Context, val any, key string, query func(val any) error) error {
	n, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return n.(Cache).TakeCtx(ctx, val, key, query)
}

// TakeWithExpire 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
func (c cluster) TakeWithExpire(val any, key string, query func(val any, expire time.Duration) error) error {
	return c.TakeWithExpireCtx(context.Background(), val, key, query)
}

// TakeWithExpireCtx 首先从缓存中获取结果，如果未找到则从DB查询并设置为给定过期时长，然后返回结果。
func (c cluster) TakeWithExpireCtx(ctx context.Context, val any, key string, query func(val any, expire time.Duration) error) error {
	n, ok := c.dispatcher.Get(key)
	if !ok {
		return c.errNotFound
	}

	return n.(Cache).TakeWithExpireCtx(ctx, val, key, query)
}
