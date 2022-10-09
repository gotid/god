package redis

import (
	"context"
	"errors"
	"fmt"
	red "github.com/go-redis/redis/v8"
	"github.com/gotid/god/lib/breaker"
	"github.com/gotid/god/lib/mapping"
	"github.com/gotid/god/lib/syncx"
	"strconv"
	"time"
)

const (
	// NodeType 意为 redis 节点。
	NodeType = "node"
	// ClusterType 意为 redis 集群。
	ClusterType = "cluster"
	// Nil 是 redis.Nil 的别称。
	Nil = red.Nil

	defaultSlowThreshold = 100 * time.Millisecond
	blockingQueryTimeout = 5 * time.Second
	readWriteTimeout     = 2 * time.Second
)

var (
	// ErrNilNode 表示一个 redis 为空节点的错误。
	ErrNilNode    = errors.New("redis 节点为空")
	slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)
)

type (
	// Option 自定义 Redis 的函数。
	Option func(r *Redis)

	// Pair 是一个用于 redis zset 的键值对。
	Pair struct {
		Member string
		Score  int64
	}

	// Redis 定义一个 redis 节点或集群。是线程安全的。
	Redis struct {
		Addr string
		Type string
		Pass string
		tls  bool
		brk  breaker.Breaker
	}

	// Node 接口表示一个 redis 节点。
	Node interface {
		red.Cmdable
	}

	// GeoLocation 用于和 GeoAdd 一起添加地理空间位置。
	GeoLocation = red.GeoLocation
	// GeoRadiusQuery 用于和 GeoRadius 一起查询地理空间索引。
	GeoRadiusQuery = red.GeoRadiusQuery
	// GeoPos 用于表示一个地理空间位置。
	GeoPos = red.GeoPos

	// Pipeliner 是 redis.Pipeliner 的别名。
	Pipeliner = red.Pipeliner

	// Z 表示排序后的集合成员。
	Z = red.Z
	// ZStore 是 redis.ZStore 的别名。
	ZStore = red.ZStore

	// IntCmd 是 redis.IntCmd 的别名。
	IntCmd = red.IntCmd
	// FloatCmd 是 redis.FloatCmd 的别名。
	FloatCmd = red.FloatCmd
	// StringCmd 是 redis.StringCmd 的别名。
	StringCmd = red.StringCmd
)

// New 返回给定地址和选项的 Redis 实例。
func New(addr string, opts ...Option) *Redis {
	r := &Redis{
		Addr: addr,
		Type: NodeType,
		brk:  breaker.New(breaker.WithName(addr)),
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

// BitCount 求 key 中比特位为 1 的数量。
func (r *Redis) BitCount(key string, start, end int64) (int64, error) {
	return r.BitCountCtx(context.Background(), key, start, end)
}

// BitCountCtx 求 key 中比特位为 1 的数量。
func (r *Redis) BitCountCtx(ctx context.Context, key string, start, end int64) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.BitCount(ctx, key, &red.BitCount{
			Start: start,
			End:   end,
		}).Result()
		return err
	}, acceptable)

	return
}

// BitOpAnd 求 keys 的逻辑并，将结果保存到 destKey。
func (r *Redis) BitOpAnd(destKey string, keys ...string) (int64, error) {
	return r.BitOpAndCtx(context.Background(), destKey, keys...)
}

// BitOpAndCtx 求 keys 的逻辑并，将结果保存到 destKey。
func (r *Redis) BitOpAndCtx(ctx context.Context, destKey string, keys ...string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.BitOpAnd(ctx, destKey, keys...).Result()
		return err
	}, acceptable)

	return
}

// BitOpOr 求 keys 的逻辑或，将结果保存到 destKey。
func (r *Redis) BitOpOr(destKey string, keys ...string) (int64, error) {
	return r.BitOpOrCtx(context.Background(), destKey, keys...)
}

// BitOpOrCtx 求 keys 的逻辑或，将结果保存到 destKey。
func (r *Redis) BitOpOrCtx(ctx context.Context, destKey string, keys ...string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.BitOpOr(ctx, destKey, keys...).Result()
		return err
	}, acceptable)

	return
}

// BitOpXor 求 keys 的逻辑异或，将结果保存到 destKey。
func (r *Redis) BitOpXor(destKey string, keys ...string) (int64, error) {
	return r.BitOpXorCtx(context.Background(), destKey, keys...)
}

// BitOpXorCtx 求 keys 的逻辑异或，将结果保存到 destKey。
func (r *Redis) BitOpXorCtx(ctx context.Context, destKey string, keys ...string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.BitOpXor(ctx, destKey, keys...).Result()
		return err
	}, acceptable)

	return
}

// BitOpNot 求 key 的逻辑非，将结果保存到 destKey。
func (r *Redis) BitOpNot(destKey, key string) (int64, error) {
	return r.BitOpNotCtx(context.Background(), destKey, key)
}

// BitOpNotCtx 求 key 的逻辑非，将结果保存到 destKey。
func (r *Redis) BitOpNotCtx(ctx context.Context, destKey, key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.BitOpNot(ctx, destKey, key).Result()
		return err
	}, acceptable)

	return
}

// BitPos 返回 key 中从 start 到 end 范围内，第一个值为 bit 的位置。
func (r *Redis) BitPos(key string, bit, start, end int64) (int64, error) {
	return r.BitPosCtx(context.Background(), key, bit, start, end)
}

// BitPosCtx 返回 key 中从 start 到 end 范围内，第一个值为 bit 的位置。
func (r *Redis) BitPosCtx(ctx context.Context, key string, bit, start, end int64) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.BitPos(ctx, key, bit, start, end).Result()
		return err
	}, acceptable)

	return
}

// BLPop 阻塞式查询节点 node 中列表 key 的第一个非空元素。否则阻塞列表至超时或有可弹元素为止。
// 注意：阻塞式查询无法获取连接池的好处（如断路器保护）。
func (r *Redis) BLPop(node Node, key string) (string, error) {
	return r.BLPopCtx(context.Background(), node, key)
}

// BLPopCtx 阻塞式查询节点 node 中列表 key 的第一个非空元素。否则阻塞列表至超时或有可弹元素为止。
// // 注意：阻塞式查询无法获取连接池的好处（如断路器保护）。
func (r *Redis) BLPopCtx(ctx context.Context, node Node, key string) (string, error) {
	return r.BLPopWithTimeoutCtx(ctx, node, blockingQueryTimeout, key)
}

// BLPopEx 阻塞式查询节点 node 中列表 key 的第一个非空元素。否则阻塞列表至超时或有可弹元素为止。
// 和 BLPop 的差异是该方法返回了一个 bool 来指示成功。
// 注意：阻塞式查询无法获取连接池的好处（如断路器保护）。
func (r *Redis) BLPopEx(node Node, key string) (string, bool, error) {
	return r.BLPopExCtx(context.Background(), node, key)
}

// BLPopExCtx 阻塞式查询节点 node 中列表 key 的第一个非空元素。否则阻塞列表至超时或有可弹元素为止。
// 和 BLPop 的差异是该方法返回了一个 bool 来指示成功。
// 注意：阻塞式查询无法获取连接池的好处（如断路器保护）。
func (r *Redis) BLPopExCtx(ctx context.Context, node Node, key string) (string, bool, error) {
	if node == nil {
		return "", false, ErrNilNode
	}

	values, err := node.BLPop(ctx, blockingQueryTimeout, key).Result()
	if err != nil {
		return "", false, err
	}

	if len(values) < 2 {
		return "", false, fmt.Errorf("列表键 %s 暂无可弹出元素", key)
	}

	return values[1], true, nil
}

// BLPopWithTimeout 阻塞式查询节点 node 中列表 key 的第一个非空元素，否则阻塞列表至超时或有可弹元素为止，可以控制阻塞时间。
// 注意：阻塞式查询无法获取连接池的好处（如断路器保护）。
func (r *Redis) BLPopWithTimeout(node Node, timeout time.Duration, key string) (string, error) {
	return r.BLPopWithTimeoutCtx(context.Background(), node, timeout, key)
}

// BLPopWithTimeoutCtx 阻塞式查询节点 node 中列表 key 的第一个非空元素，否则阻塞列表至超时或有可弹元素为止，可以控制阻塞时间。
// // 注意：阻塞式查询无法获取连接池的好处（如断路器保护）。
func (r *Redis) BLPopWithTimeoutCtx(ctx context.Context, node Node, timeout time.Duration,
	key string) (string, error) {
	if node == nil {
		return "", ErrNilNode
	}

	values, err := node.BLPop(ctx, timeout, key).Result()
	if err != nil {
		return "", err
	}

	if len(values) < 2 {
		return "", fmt.Errorf("列表键 %s 暂无可弹出元素", key)
	}

	return values[1], nil
}

// Decr 将 key 中储存的数值减1。
func (r *Redis) Decr(key string) (int64, error) {
	return r.DecrCtx(context.Background(), key)
}

// DecrCtx 将 key 中储存的数值减1。
func (r *Redis) DecrCtx(ctx context.Context, key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.Decr(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// DecrBy 将 key 中存储的数值减去 decrement。
func (r *Redis) DecrBy(key string, decrement int64) (int64, error) {
	return r.DecrByCtx(context.Background(), key, decrement)
}

// DecrByCtx 将 key 中存储的数值减去 decrement。
func (r *Redis) DecrByCtx(ctx context.Context, key string, decrement int64) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.DecrBy(ctx, key, decrement).Result()
		return err
	}, acceptable)

	return
}

// Del 删除 keys。
func (r *Redis) Del(keys ...string) (int, error) {
	return r.DelCtx(context.Background(), keys...)
}

// DelCtx 删除 keys。
func (r *Redis) DelCtx(ctx context.Context, keys ...string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.Del(ctx, keys...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// Eval 对 Lua 脚本及键值参数 keys, args 求值。
func (r *Redis) Eval(script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.EvalCtx(context.Background(), script, keys, args...)
}

// EvalCtx 对 Lua 脚本及键值参数 keys, args 求值。
func (r *Redis) EvalCtx(ctx context.Context, script string, keys []string,
	args ...interface{}) (val interface{}, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.Eval(ctx, script, keys, args...).Result()
		return err
	}, acceptable)

	return
}

// EvalSha 根据给定的 sha1 校验码，对缓存在服务器中的脚本进行求值。
func (r *Redis) EvalSha(sha string, keys []string, args ...interface{}) (interface{}, error) {
	return r.EvalShaCtx(context.Background(), sha, keys, args...)
}

// EvalShaCtx 根据给定的 sha1 校验码，对缓存在服务器中的脚本进行求值。
func (r *Redis) EvalShaCtx(ctx context.Context, sha string, keys []string,
	args ...interface{}) (val interface{}, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.EvalSha(ctx, sha, keys, args...).Result()
		return err
	}, acceptable)

	return
}

// Exists 检查 key 是否存在。
func (r *Redis) Exists(key string) (bool, error) {
	return r.ExistsCtx(context.Background(), key)
}

// ExistsCtx 检查 key 是否存在。
func (r *Redis) ExistsCtx(ctx context.Context, key string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.Exists(ctx, key).Result()
		if err != nil {
			return err
		}

		val = v == 1
		return nil
	}, acceptable)

	return
}

// Expire 设置 key 的存活秒数，过期会自动删除。
func (r *Redis) Expire(key string, seconds int) error {
	return r.ExpireCtx(context.Background(), key, seconds)
}

// ExpireCtx 设置 key 的存活秒数，过期会自动删除。
func (r *Redis) ExpireCtx(ctx context.Context, key string, seconds int) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		return node.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
	}, acceptable)
}

// ExpireAt 设置 key 的过期时间，过期会自动删除。
func (r *Redis) ExpireAt(key string, expireTime int64) error {
	return r.ExpireAtCtx(context.Background(), key, expireTime)
}

// ExpireAtCtx 设置 key 的过期时间，过期会自动删除。
func (r *Redis) ExpireAtCtx(ctx context.Context, key string, expireTime int64) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		return node.ExpireAt(ctx, key, time.Unix(expireTime, 0)).Err()
	}, acceptable)
}

// GeoAdd 添加地理位置的坐标。
func (r *Redis) GeoAdd(key string, geoLocation ...*GeoLocation) (int64, error) {
	return r.GeoAddCtx(context.Background(), key, geoLocation...)
}

// GeoAddCtx 添加地理位置的坐标。
func (r *Redis) GeoAddCtx(ctx context.Context, key string, geoLocation ...*GeoLocation) (
	val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.GeoAdd(ctx, key, geoLocation...).Result()
		if err != nil {
			return err
		}

		val = v
		return nil
	}, acceptable)

	return
}

// GeoDist 返回两个给定位置之间的距离。
// unit 可选值为：[m,km,mi,ft]，分别意为米、千米、英里、英尺。
func (r *Redis) GeoDist(key, member1, member2, unit string) (float64, error) {
	return r.GeoDistCtx(context.Background(), key, member1, member2, unit)
}

// GeoDistCtx 返回两个给定位置之间的距离。
// unit 可选值为：[m,km,mi,ft]，分别意为米、千米、英里、英尺。
func (r *Redis) GeoDistCtx(ctx context.Context, key, member1, member2, unit string) (
	val float64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.GeoDist(ctx, key, member1, member2, unit).Result()
		if err != nil {
			return err
		}

		val = v
		return nil
	}, acceptable)

	return
}

// GeoHash 返回 key 中多个成员的 Geohash 表示形式。
func (r *Redis) GeoHash(key string, members ...string) ([]string, error) {
	return r.GeoHashCtx(context.Background(), key, members...)
}

// GeoHashCtx 返回 key 中多个成员的 Geohash 表示形式。
func (r *Redis) GeoHashCtx(ctx context.Context, key string, members ...string) (
	val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.GeoHash(ctx, key, members...).Result()
		if err != nil {
			return err
		}

		val = v
		return nil
	}, acceptable)

	return
}

// GeoRadius 返回 key 中给定经纬度为中心、半径在 query 内的地理位置。
func (r *Redis) GeoRadius(key string, longitude, latitude float64, query *GeoRadiusQuery) (
	[]GeoLocation, error) {
	return r.GeoRadiusCtx(context.Background(), key, longitude, latitude, query)
}

// GeoRadiusCtx 返回 key 中以给定经纬度为中心、半径在 query 内的地理位置。
func (r *Redis) GeoRadiusCtx(ctx context.Context, key string, longitude, latitude float64,
	query *GeoRadiusQuery) (val []GeoLocation, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.GeoRadius(ctx, key, longitude, latitude, query).Result()
		if err != nil {
			return err
		}

		val = v
		return nil
	}, acceptable)

	return
}

// GeoRadiusByMember 获取 key 中以成员 member 为中心、半径在 query 内的地理位置。
func (r *Redis) GeoRadiusByMember(key, member string, query *GeoRadiusQuery) ([]GeoLocation, error) {
	return r.GeoRadiusByMemberCtx(context.Background(), key, member, query)
}

// GeoRadiusByMemberCtx 获取 key 中以成员 member 为中心、半径在 query 内的地理位置。
func (r *Redis) GeoRadiusByMemberCtx(ctx context.Context, key, member string,
	query *GeoRadiusQuery) (val []GeoLocation, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.GeoRadiusByMember(ctx, key, member, query).Result()
		if err != nil {
			return err
		}

		val = v
		return nil
	}, acceptable)

	return
}

// GeoPos 获取 key 中给定成员 members 的坐标。
func (r *Redis) GeoPos(key string, members ...string) ([]*GeoPos, error) {
	return r.GeoPosCtx(context.Background(), key, members...)
}

// GeoPosCtx 获取 key 中给定成员 members 的坐标。
func (r *Redis) GeoPosCtx(ctx context.Context, key string, members ...string) (
	val []*GeoPos, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.GeoPos(ctx, key, members...).Result()
		if err != nil {
			return err
		}

		val = v
		return nil
	}, acceptable)

	return
}

// Get 获取 key 的值。
func (r *Redis) Get(key string) (string, error) {
	return r.GetCtx(context.Background(), key)
}

// GetCtx 获取 key 的值。
func (r *Redis) GetCtx(ctx context.Context, key string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		if val, err = node.Get(ctx, key).Result(); err == red.Nil {
			return nil
		} else if err != nil {
			return err
		} else {
			return nil
		}
	}, acceptable)

	return
}

// GetBit 获取 key 上偏移量为 offset 的比特值。
func (r *Redis) GetBit(key string, offset int64) (int, error) {
	return r.GetBitCtx(context.Background(), key, offset)
}

// GetBitCtx 获取 key 上偏移量为 offset 的比特值。
func (r *Redis) GetBitCtx(ctx context.Context, key string, offset int64) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.GetBit(ctx, key, offset).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// GetSet 设置 key 的新值为 value，并返回就值。
func (r *Redis) GetSet(key, value string) (string, error) {
	return r.GetSetCtx(context.Background(), key, value)
}

// GetSetCtx 设置 key 的新值为 value，并返回就值。
func (r *Redis) GetSetCtx(ctx context.Context, key, value string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		if val, err = node.GetSet(ctx, key, value).Result(); err == red.Nil {
			return nil
		}

		return err
	}, acceptable)

	return
}

// HDel 删除哈希 key 中的多个字段 fields。
func (r *Redis) HDel(key string, fields ...string) (bool, error) {
	return r.HDelCtx(context.Background(), key, fields...)
}

// HDelCtx 删除哈希 key 中的多个字段 fields。
func (r *Redis) HDelCtx(ctx context.Context, key string, fields ...string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.HDel(ctx, key, fields...).Result()
		if err != nil {
			return err
		}

		val = v >= 1
		return nil
	}, acceptable)

	return
}

// HExists 判断哈希 key 中成员 field 是否存在。
func (r *Redis) HExists(key, field string) (bool, error) {
	return r.HExistsCtx(context.Background(), key, field)
}

// HExistsCtx 判断哈希 key 中成员 field 是否存在。
func (r *Redis) HExistsCtx(ctx context.Context, key, field string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.HExists(ctx, key, field).Result()
		return err
	}, acceptable)

	return
}

// HGet 获取哈希 key 中字段 field 的值。
func (r *Redis) HGet(key, field string) (string, error) {
	return r.HGetCtx(context.Background(), key, field)
}

// HGetCtx 获取哈希 key 中字段 field 的值。
func (r *Redis) HGetCtx(ctx context.Context, key, field string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.HGet(ctx, key, field).Result()
		return err
	}, acceptable)

	return
}

// HGetAll 获取哈希 key 的所有字段:值映射。
func (r *Redis) HGetAll(key string) (map[string]string, error) {
	return r.HGetAllCtx(context.Background(), key)
}

// HGetAllCtx 获取哈希 key 的所有字段:值映射。
func (r *Redis) HGetAllCtx(ctx context.Context, key string) (val map[string]string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.HGetAll(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// HIncrBy 为哈希 key 的字段 field 的值增加 increment。
func (r *Redis) HIncrBy(key, field string, increment int) (int, error) {
	return r.HIncrByCtx(context.Background(), key, field, increment)
}

// HIncrByCtx 为哈希 key 的字段 field 的值增加 increment。
func (r *Redis) HIncrByCtx(ctx context.Context, key, field string, increment int) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.HIncrBy(ctx, key, field, int64(increment)).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// HKeys 返回哈希 key 的所有字段。
func (r *Redis) HKeys(key string) ([]string, error) {
	return r.HKeysCtx(context.Background(), key)
}

// HKeysCtx 返回哈希 key 的所有字段。
func (r *Redis) HKeysCtx(ctx context.Context, key string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.HKeys(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// HLen 返回哈希 key 的字段数量。
func (r *Redis) HLen(key string) (int, error) {
	return r.HLenCtx(context.Background(), key)
}

// HLenCtx 返回哈希 key 的字段数量。
func (r *Redis) HLenCtx(ctx context.Context, key string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.HLen(ctx, key).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// HMGet 获取哈希 key 中所有给定字段 fields 的值。
func (r *Redis) HMGet(key string, fields ...string) ([]string, error) {
	return r.HMGetCtx(context.Background(), key, fields...)
}

// HMGetCtx 获取哈希 key 中所有给定字段 fields 的值。
func (r *Redis) HMGetCtx(ctx context.Context, key string, fields ...string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.HMGet(ctx, key, fields...).Result()
		if err != nil {
			return err
		}

		val = toStrings(v)
		return nil
	}, acceptable)

	return
}

// HSet 设置哈希 key 的字段 field 的值为 value。
func (r *Redis) HSet(key, field, value string) error {
	return r.HSetCtx(context.Background(), key, field, value)
}

// HSetCtx 设置哈希 key 的字段 field 的值为 value。
func (r *Redis) HSetCtx(ctx context.Context, key, field, value string) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		return node.HSet(ctx, key, field, value).Err()
	}, acceptable)
}

// HSetNX 当哈希 key 中字段 field 不存在时，增加字段值 field:value。
func (r *Redis) HSetNX(key, field, value string) (bool, error) {
	return r.HSetNXCtx(context.Background(), key, field, value)
}

// HSetNXCtx 当哈希 key 中字段 field 不存在时，增加字段值 field:value。
func (r *Redis) HSetNXCtx(ctx context.Context, key, field, value string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.HSetNX(ctx, key, field, value).Result()
		return err
	}, acceptable)

	return
}

// HMSet 同时设置多个键值到哈希 key。
func (r *Redis) HMSet(key string, fieldsAndValues map[string]string) error {
	return r.HMSetCtx(context.Background(), key, fieldsAndValues)
}

// HMSetCtx 同时设置多个键值到哈希 key。
func (r *Redis) HMSetCtx(ctx context.Context, key string, fieldsAndValues map[string]string) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		vals := make(map[string]interface{}, len(fieldsAndValues))
		for k, v := range fieldsAndValues {
			vals[k] = v
		}

		return node.HMSet(ctx, key, vals).Err()
	}, acceptable)
}

// HScan 迭代哈希表中的键值对。
func (r *Redis) HScan(key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	return r.HScanCtx(context.Background(), key, cursor, match, count)
}

// HScanCtx 迭代哈希表中的键值对。
func (r *Redis) HScanCtx(ctx context.Context, key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		keys, cur, err = node.HScan(ctx, key, cursor, match, count).Result()
		return err
	}, acceptable)

	return
}

// HVals 获取哈希表中所有值。
func (r *Redis) HVals(key string) ([]string, error) {
	return r.HValsCtx(context.Background(), key)
}

// HValsCtx 获取哈希表中所有值。
func (r *Redis) HValsCtx(ctx context.Context, key string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.HVals(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// Incr 将 key 中储存的数字值增一。
func (r *Redis) Incr(key string) (int64, error) {
	return r.IncrCtx(context.Background(), key)
}

// IncrCtx 将 key 中储存的数字值增一。
func (r *Redis) IncrCtx(ctx context.Context, key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.Incr(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// IncrBy 将 key 所储存的值加上给定的增量值（increment） 。
func (r *Redis) IncrBy(key string, increment int64) (int64, error) {
	return r.IncrByCtx(context.Background(), key, increment)
}

// IncrByCtx 将 key 所储存的值加上给定的增量值（increment） 。
func (r *Redis) IncrByCtx(ctx context.Context, key string, increment int64) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.IncrBy(ctx, key, increment).Result()
		return err
	}, acceptable)

	return
}

// Keys 查找所有符合给定模式( pattern)的 key 。
func (r *Redis) Keys(pattern string) ([]string, error) {
	return r.KeysCtx(context.Background(), pattern)
}

// KeysCtx 查找所有符合给定模式( pattern)的 key 。
func (r *Redis) KeysCtx(ctx context.Context, pattern string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.Keys(ctx, pattern).Result()
		return err
	}, acceptable)

	return
}

// LLen 获取列表长度。
func (r *Redis) LLen(key string) (int, error) {
	return r.LLenCtx(context.Background(), key)
}

// LLenCtx 获取列表长度。
func (r *Redis) LLenCtx(ctx context.Context, key string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.LLen(ctx, key).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// LIndex 通过索引获取列表中的元素。
func (r *Redis) LIndex(key string, index int64) (string, error) {
	return r.LIndexCtx(context.Background(), key, index)
}

// LIndexCtx 通过索引获取列表中的元素。
func (r *Redis) LIndexCtx(ctx context.Context, key string, index int64) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.LIndex(ctx, key, index).Result()
		return err
	}, acceptable)

	return
}

// LPop 移除并返回列表的第一个元素。
func (r *Redis) LPop(key string) (string, error) {
	return r.LPopCtx(context.Background(), key)
}

// LPopCtx 移除并返回列表的第一个元素。
func (r *Redis) LPopCtx(ctx context.Context, key string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.LPop(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// LPush 将一个或多个值插入到列表头部。
func (r *Redis) LPush(key string, values ...interface{}) (int, error) {
	return r.LPushCtx(context.Background(), key, values...)
}

// LPushCtx 将一个或多个值插入到列表头部。
func (r *Redis) LPushCtx(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.LPush(ctx, key, values...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// LRange 获取列表指定范围内的元素。
func (r *Redis) LRange(key string, start, stop int) ([]string, error) {
	return r.LRangeCtx(context.Background(), key, start, stop)
}

// LRangeCtx 获取列表指定范围内的元素。
func (r *Redis) LRangeCtx(ctx context.Context, key string, start, stop int) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.LRange(ctx, key, int64(start), int64(stop)).Result()
		return err
	}, acceptable)

	return
}

// LRem 移除列表元素。
// count 控制移除的方向和数量： 为正数从头到尾找，为负数从尾到头找，都删除count个；为0移除所有。
func (r *Redis) LRem(key string, count int, value string) (int, error) {
	return r.LRemCtx(context.Background(), key, count, value)
}

// LRemCtx 移除列表元素。
// count 控制移除的方向和数量： 为正数从头到尾找，为负数从尾到头找，都删除count个；为0移除所有。
func (r *Redis) LRemCtx(ctx context.Context, key string, count int, value string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.LRem(ctx, key, int64(count), value).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// LTrim 修剪列表，只保留指定起止区间的元素。
func (r *Redis) LTrim(key string, start, stop int64) error {
	return r.LTrimCtx(context.Background(), key, start, stop)
}

// LTrimCtx 修剪列表，只保留指定起止区间的元素。
func (r *Redis) LTrimCtx(ctx context.Context, key string, start, stop int64) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		return node.LTrim(ctx, key, start, stop).Err()
	}, acceptable)
}

// MGet 获取所有给定 key 的值。
func (r *Redis) MGet(keys ...string) ([]string, error) {
	return r.MGetCtx(context.Background(), keys...)
}

// MGetCtx 获取所有给定 key 的值。
func (r *Redis) MGetCtx(ctx context.Context, keys ...string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.MGet(ctx, keys...).Result()
		if err != nil {
			return err
		}

		val = toStrings(v)
		return nil
	}, acceptable)

	return
}

// Persist 移除 key 的过期时间，key 将持久保持。
func (r *Redis) Persist(key string) (bool, error) {
	return r.PersistCtx(context.Background(), key)
}

// PersistCtx 移除 key 的过期时间，key 将持久保持。
func (r *Redis) PersistCtx(ctx context.Context, key string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.Persist(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// PFAdd 将 values 加入到键为 key 的 HyperLogLog 中，用于快速统计超大数据的唯一元素（基数）估算值。
func (r *Redis) PFAdd(key string, values ...interface{}) (bool, error) {
	return r.PFAddCtx(context.Background(), key, values...)
}

// PFAddCtx 将 values 加入到键为 key 的 HyperLogLog 中，用于快速统计超大数据的唯一元素（基数）估算值。
func (r *Redis) PFAddCtx(ctx context.Context, key string, values ...interface{}) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.PFAdd(ctx, key, values...).Result()
		if err != nil {
			return err
		}

		val = v >= 1
		return nil
	}, acceptable)

	return
}

// PFCount 返回给定 HyperLogLog 的基数估算值。
func (r *Redis) PFCount(key string) (int64, error) {
	return r.PFCountCtx(context.Background(), key)
}

// PFCountCtx 返回给定 HyperLogLog 的基数估算值。
func (r *Redis) PFCountCtx(ctx context.Context, key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.PFCount(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// PFMerge 将多个 HyperLogLog 合并为一个 HyperLogLog。
func (r *Redis) PFMerge(dest string, keys ...string) error {
	return r.PFMergeCtx(context.Background(), dest, keys...)
}

// PFMergeCtx 将多个 HyperLogLog 合并为一个 HyperLogLog。
func (r *Redis) PFMergeCtx(ctx context.Context, dest string, keys ...string) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		_, err = node.PFMerge(ctx, dest, keys...).Result()
		return err
	}, acceptable)
}

// Ping 检测 redis 服务是否启动。
func (r *Redis) Ping() bool {
	return r.PingCtx(context.Background())
}

// PingCtx 检测 redis 服务是否启动。
func (r *Redis) PingCtx(ctx context.Context) (val bool) {
	// 忽略错误，错误意为未启动
	_ = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			val = false
			return nil
		}

		v, err := node.Ping(ctx).Result()
		if err != nil {
			val = false
			return nil
		}

		val = v == "PONG"
		return nil
	}, acceptable)

	return
}

// Pipelined 执行管道化函数。
func (r *Redis) Pipelined(fn func(Pipeliner) error) error {
	return r.PipelinedCtx(context.Background(), fn)
}

// PipelinedCtx 执行管道化函数。
// 结果需要调用 Pipeline.Exec() 来获取。
func (r *Redis) PipelinedCtx(ctx context.Context, fn func(Pipeliner) error) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		_, err = node.Pipelined(ctx, fn)
		return err
	}, acceptable)
}

// RPop 移除并返回列表的最后一个元素。
func (r *Redis) RPop(key string) (string, error) {
	return r.RPopCtx(context.Background(), key)
}

// RPopCtx 移除并返回列表的最后一个元素。
func (r *Redis) RPopCtx(ctx context.Context, key string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.RPop(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// RPush 在列表右侧添加一个或多个值。
func (r *Redis) RPush(key string, values ...interface{}) (int, error) {
	return r.RPushCtx(context.Background(), key, values...)
}

// RPushCtx 在列表右侧添加一个或多个值。
func (r *Redis) RPushCtx(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.RPush(ctx, key, values...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// SAdd 向集合添加一个或多个成员。
func (r *Redis) SAdd(key string, values ...interface{}) (int, error) {
	return r.SAddCtx(context.Background(), key, values...)
}

// SAddCtx 向集合添加一个或多个成员。
func (r *Redis) SAddCtx(ctx context.Context, key string, values ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.SAdd(ctx, key, values...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// Scan 迭代数据库中的数据库键。
func (r *Redis) Scan(cursor uint64, match string, count int64) (keys []string, cur uint64, err error) {
	return r.ScanCtx(context.Background(), cursor, match, count)
}

// ScanCtx 迭代数据库中的数据库键。
func (r *Redis) ScanCtx(ctx context.Context, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		keys, cur, err = node.Scan(ctx, cursor, match, count).Result()
		return err
	}, acceptable)

	return
}

// SetBit 设置或清除 key 上偏移量为 offset 的比特值为 value。
func (r *Redis) SetBit(key string, offset int64, value int) (int, error) {
	return r.SetBitCtx(context.Background(), key, offset, value)
}

// SetBitCtx 设置或清除 key 上偏移量为 offset 的比特值为 value。
func (r *Redis) SetBitCtx(ctx context.Context, key string, offset int64, value int) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.SetBit(ctx, key, offset, value).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// SScan 迭代集合中的元素。
func (r *Redis) SScan(key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	return r.SScanCtx(context.Background(), key, cursor, match, count)
}

// SScanCtx 迭代集合中的元素。
func (r *Redis) SScanCtx(ctx context.Context, key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		keys, cur, err = node.SScan(ctx, key, cursor, match, count).Result()
		return err
	}, acceptable)

	return
}

// SCard 获取集合的成员数。
func (r *Redis) SCard(key string) (int64, error) {
	return r.SCardCtx(context.Background(), key)
}

// SCardCtx 获取集合的成员数。
func (r *Redis) SCardCtx(ctx context.Context, key string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SCard(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// ScriptLoad 将脚本 script 添加到脚本缓存中，但并不立即执行这个脚本。
// 返回脚本的 sha1 校验码。
func (r *Redis) ScriptLoad(script string) (string, error) {
	return r.ScriptLoadCtx(context.Background(), script)
}

// ScriptLoadCtx 将脚本 script 添加到脚本缓存中，但并不立即执行这个脚本。
// 返回脚本的 sha1 校验码。
func (r *Redis) ScriptLoadCtx(ctx context.Context, script string) (string, error) {
	node, err := getRedis(r)
	if err != nil {
		return "", err
	}

	return node.ScriptLoad(ctx, script).Result()
}

// Set 设置 key 的值。
func (r *Redis) Set(key, value string) error {
	return r.SetCtx(context.Background(), key, value)
}

// SetCtx 设置 key 的值。
func (r *Redis) SetCtx(ctx context.Context, key, value string) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		return node.Set(ctx, key, value, 0).Err()
	}, acceptable)
}

// SetEx 设置键值及其存活秒数，过期会自动删除。
func (r *Redis) SetEx(key, value string, seconds int) error {
	return r.SetExCtx(context.Background(), key, value, seconds)
}

// SetExCtx 设置键值及其存活秒数，过期会自动删除。
func (r *Redis) SetExCtx(ctx context.Context, key, value string, seconds int) error {
	return r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		return node.Set(ctx, key, value, time.Duration(seconds)*time.Second).Err()
	}, acceptable)
}

// SetNX 当 key 不存在时，设置键值对。
func (r *Redis) SetNX(key, value string) (bool, error) {
	return r.SetNXCtx(context.Background(), key, value)
}

// SetNXCtx 当 key 不存在时，设置键值对。
func (r *Redis) SetNXCtx(ctx context.Context, key, value string) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SetNX(ctx, key, value, 0).Result()
		return err
	}, acceptable)

	return
}

// SetNXEx 当 key 不存时，设置键值对及其存活秒数，过期会自动删除。
func (r *Redis) SetNXEx(key, value string, seconds int) (bool, error) {
	return r.SetNXExCtx(context.Background(), key, value, seconds)
}

// SetNXExCtx 当 key 不存时，设置键值对及其存活秒数，过期会自动删除。
func (r *Redis) SetNXExCtx(ctx context.Context, key, value string, seconds int) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SetNX(ctx, key, value, time.Duration(seconds)*time.Second).Result()
		return err
	}, acceptable)

	return
}

// SIsMember 判断 member 是否为集合 key 的成员。
func (r *Redis) SIsMember(key string, member interface{}) (bool, error) {
	return r.SIsMemberCtx(context.Background(), key, member)
}

// SIsMemberCtx 判断 member 是否为集合 key 的成员。
func (r *Redis) SIsMemberCtx(ctx context.Context, key string, member interface{}) (val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SIsMember(ctx, key, member).Result()
		return err
	}, acceptable)

	return
}

// SMembers 返回集合 key 中的所有成员。
func (r *Redis) SMembers(key string) ([]string, error) {
	return r.SMembersCtx(context.Background(), key)
}

// SMembersCtx 返回集合 key 中的所有成员。
func (r *Redis) SMembersCtx(ctx context.Context, key string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SMembers(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// SPop 随机移除并返回集合 key 中的一个成员。
func (r *Redis) SPop(key string) (string, error) {
	return r.SPopCtx(context.Background(), key)
}

// SPopCtx 随机移除并返回集合 key 中的一个成员。
func (r *Redis) SPopCtx(ctx context.Context, key string) (val string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SPop(ctx, key).Result()
		return err
	}, acceptable)

	return
}

// SRandMember 随机返回集合 key 中的 count 个成员。
func (r *Redis) SRandMember(key string, count int) ([]string, error) {
	return r.SRandMemberCtx(context.Background(), key, count)
}

// SRandMemberCtx 随机返回集合 key 中的 count 个成员。
func (r *Redis) SRandMemberCtx(ctx context.Context, key string, count int) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SRandMemberN(ctx, key, int64(count)).Result()
		return err
	}, acceptable)

	return
}

// SRem 移除集合 key 中的一个或多个成员 members。
func (r *Redis) SRem(key string, members ...interface{}) (int, error) {
	return r.SRemCtx(context.Background(), key, members...)
}

// SRemCtx 移除集合 key 中的一个或多个成员 members。
func (r *Redis) SRemCtx(ctx context.Context, key string, members ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.SRem(ctx, key, members...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// String 返回 r 的字符串表示形式。
func (r *Redis) String() string {
	return r.Addr
}

// SUnion 返回一组集合 keys 的并集。
func (r *Redis) SUnion(keys ...string) ([]string, error) {
	return r.SUnionCtx(context.Background(), keys...)
}

// SUnionCtx 返回一组集合 keys 的并集。
func (r *Redis) SUnionCtx(ctx context.Context, keys ...string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SUnion(ctx, keys...).Result()
		return err
	}, acceptable)

	return
}

// SUnionStore 求一组集合 keys 的并集，将结果存储在 destination 集合中。
func (r *Redis) SUnionStore(destination string, keys ...string) (int, error) {
	return r.SUnionStoreCtx(context.Background(), destination, keys...)
}

// SUnionStoreCtx 求一组集合 keys 的并集，将结果存储在 destination 集合中。
func (r *Redis) SUnionStoreCtx(ctx context.Context, destination string, keys ...string) (
	val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.SUnionStore(ctx, destination, keys...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// SDiff 返回第一个集合相比于其他集合独有的成员。
func (r *Redis) SDiff(keys ...string) ([]string, error) {
	return r.SDiffCtx(context.Background(), keys...)
}

// SDiffCtx 返回第一个集合相比于其他集合独有的成员。
func (r *Redis) SDiffCtx(ctx context.Context, keys ...string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SDiff(ctx, keys...).Result()
		return err
	}, acceptable)

	return
}

// SDiffStore 返回第一个集合相比于其他集合独有的成员，将结果存储到结合 destination。
func (r *Redis) SDiffStore(destination string, keys ...string) (int, error) {
	return r.SDiffStoreCtx(context.Background(), destination, keys...)
}

// SDiffStoreCtx 返回第一个集合相比于其他集合独有的成员，将结果存储到结合 destination。
func (r *Redis) SDiffStoreCtx(ctx context.Context, destination string, keys ...string) (
	val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.SDiffStore(ctx, destination, keys...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// SInter 返回所有集合 keys 的交集。
func (r *Redis) SInter(keys ...string) ([]string, error) {
	return r.SInterCtx(context.Background(), keys...)
}

// SInterCtx 返回所有集合 keys 的交集。
func (r *Redis) SInterCtx(ctx context.Context, keys ...string) (val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.SInter(ctx, keys...).Result()
		return err
	}, acceptable)

	return
}

// SInterStore 返回所有集合 keys 的交集，将结果存储至集合 destination。
func (r *Redis) SInterStore(destination string, keys ...string) (int, error) {
	return r.SInterStoreCtx(context.Background(), destination, keys...)
}

// SInterStoreCtx 返回所有集合 keys 的交集，将结果存储至集合 destination。
func (r *Redis) SInterStoreCtx(ctx context.Context, destination string, keys ...string) (
	val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.SInterStore(ctx, destination, keys...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// TTL 返回 key 的剩余生存秒数。
func (r *Redis) TTL(key string) (int, error) {
	return r.TTLCtx(context.Background(), key)
}

// TTLCtx 返回 key 的剩余生存秒数。
func (r *Redis) TTLCtx(ctx context.Context, key string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		duration, err := node.TTL(ctx, key).Result()
		if err != nil {
			return err
		}

		val = int(duration / time.Second)
		return nil
	}, acceptable)

	return
}

// ZAdd 向有序集合 key 添加或更新一个成员及其分数。
func (r *Redis) ZAdd(key string, score int64, member string) (bool, error) {
	return r.ZAddCtx(context.Background(), key, score, member)
}

// ZAddFloat 向有序集合 key 添加或更新一个成员及其分数。
func (r *Redis) ZAddFloat(key string, score float64, member string) (bool, error) {
	return r.ZAddFloatCtx(context.Background(), key, score, member)
}

// ZAddCtx 向有序集合 key 添加或更新一个成员及其分数。
func (r *Redis) ZAddCtx(ctx context.Context, key string, score int64, member string) (
	val bool, err error) {
	return r.ZAddFloatCtx(ctx, key, float64(score), member)
}

// ZAddFloatCtx 向有序集合 key 添加或更新一个成员及其分数。
func (r *Redis) ZAddFloatCtx(ctx context.Context, key string, score float64, member string) (
	val bool, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZAdd(ctx, key, &red.Z{
			Score:  score,
			Member: member,
		}).Result()
		if err != nil {
			return err
		}

		val = v == 1
		return nil
	}, acceptable)

	return
}

// ZAdds 向有序集合添加或更新一个或多个成员及其分数。
func (r *Redis) ZAdds(key string, ps ...Pair) (int64, error) {
	return r.ZAddsCtx(context.Background(), key, ps...)
}

// ZAddsCtx 向有序集合添加或更新一个或多个成员及其分数。
func (r *Redis) ZAddsCtx(ctx context.Context, key string, ps ...Pair) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		var zs []*red.Z
		for _, p := range ps {
			z := &red.Z{Score: float64(p.Score), Member: p.Member}
			zs = append(zs, z)
		}

		v, err := node.ZAdd(ctx, key, zs...).Result()
		if err != nil {
			return err
		}

		val = v
		return nil
	}, acceptable)

	return
}

// ZCard 获取有序集合 key 的成员数量。
func (r *Redis) ZCard(key string) (int, error) {
	return r.ZCardCtx(context.Background(), key)
}

// ZCardCtx 获取有序集合 key 的成员数量。
func (r *Redis) ZCardCtx(ctx context.Context, key string) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZCard(ctx, key).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZCount 求有序集合中指定分数区间的成员数量。
func (r *Redis) ZCount(key string, start, stop int64) (int, error) {
	return r.ZCountCtx(context.Background(), key, start, stop)
}

// ZCountCtx 求有序集合中指定分数区间的成员数量。
func (r *Redis) ZCountCtx(ctx context.Context, key string, start, stop int64) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZCount(ctx, key, strconv.FormatInt(start, 10),
			strconv.FormatInt(stop, 10)).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZIncrBy 在有序集合中指定成员的分数上增加 increment 分。
func (r *Redis) ZIncrBy(key string, increment int64, member string) (int64, error) {
	return r.ZIncrByCtx(context.Background(), key, increment, member)
}

// ZIncrByCtx 在有序集合中指定成员的分数上增加 increment 分。
func (r *Redis) ZIncrByCtx(ctx context.Context, key string, increment int64, member string) (
	val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZIncrBy(ctx, key, float64(increment), member).Result()
		if err != nil {
			return err
		}

		val = int64(v)
		return nil
	}, acceptable)

	return
}

// ZScore 返回有序集合中指定成员的分数。
func (r *Redis) ZScore(key, value string) (int64, error) {
	return r.ZScoreCtx(context.Background(), key, value)
}

// ZScoreCtx 返回有序集合中指定成员的分数。
func (r *Redis) ZScoreCtx(ctx context.Context, key, member string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZScore(ctx, key, member).Result()
		if err != nil {
			return err
		}

		val = int64(v)
		return nil
	}, acceptable)

	return
}

// ZRank 获取有序集合中给定成员的升序索引排名。
func (r *Redis) ZRank(key, field string) (int64, error) {
	return r.ZRankCtx(context.Background(), key, field)
}

// ZRankCtx 获取有序集合中给定成员的升序索引排名。
func (r *Redis) ZRankCtx(ctx context.Context, key, field string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.ZRank(ctx, key, field).Result()
		return err
	}, acceptable)

	return
}

// ZRem 移除有序集合中给定的一个或多个成员。
func (r *Redis) ZRem(key string, members ...interface{}) (int, error) {
	return r.ZRemCtx(context.Background(), key, members...)
}

// ZRemCtx 移除有序集合中给定的一个或多个成员。
func (r *Redis) ZRemCtx(ctx context.Context, key string, members ...interface{}) (val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRem(ctx, key, members...).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZRemRangeByScore 移除有序集合中给定分数区间的所有成员。
func (r *Redis) ZRemRangeByScore(key string, start, stop int64) (int, error) {
	return r.ZRemRangeByScoreCtx(context.Background(), key, start, stop)
}

// ZRemRangeByScoreCtx 移除有序集合中给定分数区间的所有成员。
func (r *Redis) ZRemRangeByScoreCtx(ctx context.Context, key string, start, stop int64) (
	val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRemRangeByScore(ctx, key, strconv.FormatInt(start, 10),
			strconv.FormatInt(stop, 10)).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZRemRangeByRank 移除有序集合中给定排名区间的所有成员。
func (r *Redis) ZRemRangeByRank(key string, start, stop int64) (int, error) {
	return r.ZRemRangeByRankCtx(context.Background(), key, start, stop)
}

// ZRemRangeByRankCtx 移除有序集合中给定排名区间的所有成员。
func (r *Redis) ZRemRangeByRankCtx(ctx context.Context, key string, start, stop int64) (
	val int, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRemRangeByRank(ctx, key, start, stop).Result()
		if err != nil {
			return err
		}

		val = int(v)
		return nil
	}, acceptable)

	return
}

// ZRange 获取有序集合中在索引区间内的所有成员。
func (r *Redis) ZRange(key string, start, stop int64) ([]string, error) {
	return r.ZRangeCtx(context.Background(), key, start, stop)
}

// ZRangeCtx 获取有序集合中在索引区间内的所有成员。
func (r *Redis) ZRangeCtx(ctx context.Context, key string, start, stop int64) (
	val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.ZRange(ctx, key, start, stop).Result()
		return err
	}, acceptable)

	return
}

// ZRangeWithScores 获取有序集合中在索引区间内的所有成员及其分数。
func (r *Redis) ZRangeWithScores(key string, start, stop int64) ([]Pair, error) {
	return r.ZRangeWithScoresCtx(context.Background(), key, start, stop)
}

// ZRangeWithScoresCtx 获取有序集合中在索引区间内的所有成员及其分数。
func (r *Redis) ZRangeWithScoresCtx(ctx context.Context, key string, start, stop int64) (
	val []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRangeWithScores(ctx, key, start, stop).Result()
		if err != nil {
			return err
		}

		val = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRevRangeWithScores 获取有序集合中在索引区间内的所有成员及其分数，按索引从高到低排序。
func (r *Redis) ZRevRangeWithScores(key string, start, stop int64) ([]Pair, error) {
	return r.ZRevRangeWithScoresCtx(context.Background(), key, start, stop)
}

// ZRevRangeWithScoresCtx 获取有序集合中在索引区间内的所有成员及其分数，按索引从高到低排序。
func (r *Redis) ZRevRangeWithScoresCtx(ctx context.Context, key string, start, stop int64) (
	val []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRevRangeWithScores(ctx, key, start, stop).Result()
		if err != nil {
			return err
		}

		val = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRangeByScoreWithScores 获取有序集合中在分数区间内的所有成员及其分数。
func (r *Redis) ZRangeByScoreWithScores(key string, start, stop int64) ([]Pair, error) {
	return r.ZRangeByScoreWithScoresCtx(context.Background(), key, start, stop)
}

// ZRangeByScoreWithScoresCtx 获取有序集合中在分数区间内的所有成员及其分数。
func (r *Redis) ZRangeByScoreWithScoresCtx(ctx context.Context, key string, start, stop int64) (
	val []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
			Min: strconv.FormatInt(start, 10),
			Max: strconv.FormatInt(stop, 10),
		}).Result()
		if err != nil {
			return err
		}

		val = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRangeByScoreWithScoresAndLimit 获取有序集合中在分数区间内的指定分页的成员及其分数。
func (r *Redis) ZRangeByScoreWithScoresAndLimit(key string, start, stop int64,
	page, size int) ([]Pair, error) {
	return r.ZRangeByScoreWithScoresAndLimitCtx(context.Background(), key, start, stop, page, size)
}

// ZRangeByScoreWithScoresAndLimitCtx 获取有序集合中在分数区间内的指定分页的成员及其分数。
func (r *Redis) ZRangeByScoreWithScoresAndLimitCtx(ctx context.Context, key string, start,
	stop int64, page, size int) (val []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		if size <= 0 {
			return nil
		}

		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
			Min:    strconv.FormatInt(start, 10),
			Max:    strconv.FormatInt(stop, 10),
			Offset: int64(page * size),
			Count:  int64(size),
		}).Result()
		if err != nil {
			return err
		}

		val = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRevRange 获取有序集合中在索引区间内的所有成员，按分数从高到低排序。
func (r *Redis) ZRevRange(key string, start, stop int64) ([]string, error) {
	return r.ZRevRangeCtx(context.Background(), key, start, stop)
}

// ZRevRangeCtx 获取有序集合中在索引区间内的所有成员，按分数从高到低排序。
func (r *Redis) ZRevRangeCtx(ctx context.Context, key string, start, stop int64) (
	val []string, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.ZRevRange(ctx, key, start, stop).Result()
		return err
	}, acceptable)

	return
}

// ZRevRangeByScoreWithScores 获取有序集合中在分数区间内的所有成员及其分数，按索引从高到低排序。
func (r *Redis) ZRevRangeByScoreWithScores(key string, start, stop int64) ([]Pair, error) {
	return r.ZRevRangeByScoreWithScoresCtx(context.Background(), key, start, stop)
}

// ZRevRangeByScoreWithScoresCtx 获取有序集合中在分数区间内的所有成员及其分数，按索引从高到低排序。
func (r *Redis) ZRevRangeByScoreWithScoresCtx(ctx context.Context, key string, start, stop int64) (
	val []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRevRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
			Min: strconv.FormatInt(start, 10),
			Max: strconv.FormatInt(stop, 10),
		}).Result()
		if err != nil {
			return err
		}

		val = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRevRangeByScoreWithScoresAndLimit 获取有序集合中在分数区间内的指定分页的成员及其分数，按分数从高到低排序。
func (r *Redis) ZRevRangeByScoreWithScoresAndLimit(key string, start, stop int64,
	page, size int) ([]Pair, error) {
	return r.ZRevRangeByScoreWithScoresAndLimitCtx(context.Background(), key, start, stop, page, size)
}

// ZRevRangeByScoreWithScoresAndLimitCtx 获取有序集合中在分数区间内的指定分页的成员及其分数，按分数从高到低排序。
func (r *Redis) ZRevRangeByScoreWithScoresAndLimitCtx(ctx context.Context, key string,
	start, stop int64, page, size int) (val []Pair, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		if size <= 0 {
			return nil
		}

		node, err := getRedis(r)
		if err != nil {
			return err
		}

		v, err := node.ZRevRangeByScoreWithScores(ctx, key, &red.ZRangeBy{
			Min:    strconv.FormatInt(start, 10),
			Max:    strconv.FormatInt(stop, 10),
			Offset: int64(page * size),
			Count:  int64(size),
		}).Result()
		if err != nil {
			return err
		}

		val = toPairs(v)
		return nil
	}, acceptable)

	return
}

// ZRevRank 获取有序集合中给定成员的降序索引排名。
func (r *Redis) ZRevRank(key, member string) (int64, error) {
	return r.ZRevRankCtx(context.Background(), key, member)
}

// ZRevRankCtx 获取有序集合中给定成员的降序索引排名。
func (r *Redis) ZRevRankCtx(ctx context.Context, key, member string) (val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.ZRevRank(ctx, key, member).Result()
		return err
	}, acceptable)

	return
}

// ZUnionStore 求一个或多个有序集合的并集，将结果存储至 dest。
func (r *Redis) ZUnionStore(dest string, store *ZStore) (int64, error) {
	return r.ZUnionStoreCtx(context.Background(), dest, store)
}

// ZUnionStoreCtx 求一个或多个有序集合的并集，将结果存储至 dest。
func (r *Redis) ZUnionStoreCtx(ctx context.Context, dest string, store *ZStore) (
	val int64, err error) {
	err = r.brk.DoWithAcceptable(func() error {
		node, err := getRedis(r)
		if err != nil {
			return err
		}

		val, err = node.ZUnionStore(ctx, dest, store).Result()
		return err
	}, acceptable)

	return
}

// SetSlowThreshold 设置慢调用时长阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}

// WithCluster 自定义 Redis 为集群模式。
func WithCluster() Option {
	return func(r *Redis) {
		r.Type = ClusterType
	}
}

// WithPass 自定义 Redis 的密码。
func WithPass(pass string) Option {
	return func(r *Redis) {
		r.Pass = pass
	}
}

// WithTLS 自定义 Redis 启用 TLS。
func WithTLS() Option {
	return func(r *Redis) {
		r.tls = true
	}
}

// 获取 redis 节点。
func getRedis(r *Redis) (Node, error) {
	switch r.Type {
	case ClusterType:
		return getCluster(r)
	case NodeType:
		return getClient(r)
	default:
		return nil, fmt.Errorf("不支持 redis 类型 '%s'", r.Type)
	}
}

// 判断错误是否可接受。
func acceptable(err error) bool {
	return err == nil || err == red.Nil || err == context.Canceled
}

// 将 redis 返回的有序集合转为 Pair
func toPairs(values []red.Z) []Pair {
	pairs := make([]Pair, len(values))
	for i, v := range values {
		switch member := v.Member.(type) {
		case string:
			pairs[i] = Pair{
				Member: member,
				Score:  int64(v.Score),
			}
		default:
			pairs[i] = Pair{
				Member: mapping.Repr(v.Member),
				Score:  int64(v.Score),
			}
		}
	}

	return pairs
}

func toStrings(values []interface{}) []string {
	ret := make([]string, len(values))
	for i, v := range values {
		if v == nil {
			ret[i] = ""
		} else {
			switch val := v.(type) {
			case string:
				ret[i] = val
			default:
				ret[i] = mapping.Repr(val)
			}
		}
	}

	return ret
}
