package kv

import (
	"context"
	"errors"
	"github.com/gotid/god/lib/errorx"
	"github.com/gotid/god/lib/hash"
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/redis"
	"log"
)

var ErrNoRedisNode = errors.New("未找到键对应的 redis 节点")

type (
	// Store 接口代表一个键值对存储 KVStore。
	Store interface {
		// Decr 将 key 中储存的数值减1。
		Decr(key string) (int64, error)
		// DecrCtx 将 key 中储存的数值减1。
		DecrCtx(ctx context.Context, key string) (int64, error)
		// DecrBy 将 key 中存储的数值减去 decrement。
		DecrBy(key string, decrement int64) (int64, error)
		// DecrByCtx 将 key 中存储的数值减去 decrement。
		DecrByCtx(ctx context.Context, key string, decrement int64) (int64, error)
		// Del 删除 keys。
		Del(keys ...string) (int, error)
		// DelCtx 删除 keys。
		DelCtx(ctx context.Context, keys ...string) (int, error)
		// Eval 对 Lua 脚本及键值参数 keys, args 求值。
		Eval(script string, key string, args ...interface{}) (interface{}, error)
		// EvalCtx 对 Lua 脚本及键值参数 keys, args 求值。
		EvalCtx(ctx context.Context, script string, key string, args ...interface{}) (interface{}, error)
		// Exists 检查 key 是否存在。
		Exists(key string) (bool, error)
		// ExistsCtx 检查 key 是否存在。
		ExistsCtx(ctx context.Context, key string) (bool, error)
		// Expire 设置 key 的存活秒数，过期会自动删除。
		Expire(key string, seconds int) error
		// ExpireCtx 设置 key 的存活秒数，过期会自动删除。
		ExpireCtx(ctx context.Context, key string, seconds int) error
		// ExpireAt 设置 key 的过期时间，过期会自动删除。
		ExpireAt(key string, expireTime int64) error
		// ExpireAtCtx 设置 key 的过期时间，过期会自动删除。
		ExpireAtCtx(ctx context.Context, key string, expireTime int64) error
		// Get 获取 key 的值。
		Get(key string) (string, error)
		// GetCtx 获取 key 的值。
		GetCtx(ctx context.Context, key string) (string, error)
		// GetSet 设置 key 的新值为 value，并返回就值。
		GetSet(key, value string) (string, error)
		// GetSetCtx 设置 key 的新值为 value，并返回就值。
		GetSetCtx(ctx context.Context, key, value string) (string, error)
		// GetBit  获取 key 上偏移量为 offset 的比特值。
		GetBit(key string, offset int64) (int, error)
		// GetBitCtx 获取 key 上偏移量为 offset 的比特值。
		GetBitCtx(ctx context.Context, key string, offset int64) (int, error)
		// HDel 删除哈希 key 中的给定字段 fields。
		HDel(key, field string) (bool, error)
		// HDelCtx 删除哈希 key 中的给定字段 fields。
		HDelCtx(ctx context.Context, key, field string) (bool, error)
		// HExists 判断哈希 key 中成员 field 是否存在。
		HExists(key, field string) (bool, error)
		// HExistsCtx 判断哈希 key 中成员 field 是否存在。
		HExistsCtx(ctx context.Context, key, field string) (bool, error)
		// HGet 获取哈希 key 中字段 field 的值。
		HGet(key, field string) (string, error)
		// HGetCtx 获取哈希 key 中字段 field 的值。
		HGetCtx(ctx context.Context, key, field string) (string, error)
		// HGetAll 获取哈希 key 的所有字段:值映射。
		HGetAll(key string) (map[string]string, error)
		// HGetAllCtx 获取哈希 key 的所有字段:值映射。
		HGetAllCtx(ctx context.Context, key string) (map[string]string, error)
		// HIncrBy 为哈希 key 的字段 field 的值增加 increment。
		HIncrBy(key, field string, increment int) (int, error)
		// HIncrByCtx 为哈希 key 的字段 field 的值增加 increment。
		HIncrByCtx(ctx context.Context, key, field string, increment int) (int, error)
		// HKeys 返回哈希 key 的所有字段。
		HKeys(key string) ([]string, error)
		// HKeysCtx 返回哈希 key 的所有字段。
		HKeysCtx(ctx context.Context, key string) ([]string, error)
		// HLen 返回哈希 key 的字段数量。
		HLen(key string) (int, error)
		// HLenCtx 返回哈希 key 的字段数量。
		HLenCtx(ctx context.Context, key string) (int, error)
		// HMGet 获取哈希 key 中所有给定字段 fields 的值。
		HMGet(key string, fields ...string) ([]string, error)
		// HMGetCtx 获取哈希 key 中所有给定字段 fields 的值。
		HMGetCtx(ctx context.Context, key string, fields ...string) ([]string, error)
		// HSet 设置哈希 key 的字段 field 的值为 value。
		HSet(key, field, value string) error
		// HSetCtx 设置哈希 key 的字段 field 的值为 value。
		HSetCtx(ctx context.Context, key, field, value string) error
		// HSetNx 当哈希 key 中字段 field 不存在时，增加字段值 field:value。
		HSetNx(key, field, value string) (bool, error)
		// HSetNxCtx 当哈希 key 中字段 field 不存在时，增加字段值 field:value。
		HSetNxCtx(ctx context.Context, key, field, value string) (bool, error)
		// HMSet 同时设置多个键值到哈希 key。
		HMSet(key string, fieldsAndValues map[string]string) error
		// HMSetCtx 同时设置多个键值到哈希 key。
		HMSetCtx(ctx context.Context, key string, fieldsAndValues map[string]string) error
		// HVals 获取哈希表中所有值。
		HVals(key string) ([]string, error)
		// HValsCtx 获取哈希表中所有值。
		HValsCtx(ctx context.Context, key string) ([]string, error)
		// Incr 将 key 中储存的数字值增一。
		Incr(key string) (int64, error)
		// IncrCtx 将 key 中储存的数字值增一。
		IncrCtx(ctx context.Context, key string) (int64, error)
		// IncrBy 将 key 所储存的值加上给定的增量值（increment） 。
		IncrBy(key string, increment int64) (int64, error)
		// IncrByCtx 将 key 所储存的值加上给定的增量值（increment） 。
		IncrByCtx(ctx context.Context, key string, increment int64) (int64, error)
		// LLen 获取列表长度。
		LLen(key string) (int, error)
		// LLenCtx 获取列表长度。
		LLenCtx(ctx context.Context, key string) (val int, err error)
		// LIndex 通过索引获取列表中的元素。
		LIndex(key string, index int64) (string, error)
		// LIndexCtx 通过索引获取列表中的元素。
		LIndexCtx(ctx context.Context, key string, index int64) (val string, err error)
		// LPop 移除并返回列表的第一个元素。
		LPop(key string) (string, error)
		// LPopCtx 移除并返回列表的第一个元素。
		LPopCtx(ctx context.Context, key string) (val string, err error)
		// LPush 将一个或多个值插入到列表头部。
		LPush(key string, values ...interface{}) (int, error)
		// LPushCtx 将一个或多个值插入到列表头部。
		LPushCtx(ctx context.Context, key string, values ...interface{}) (val int, err error)
		// LRange 获取列表指定范围内的元素。
		LRange(key string, start, stop int) ([]string, error)
		// LRangeCtx 获取列表指定范围内的元素。
		LRangeCtx(ctx context.Context, key string, start, stop int) (val []string, err error)
		// LRem 移除列表元素。count 控制移除的方向和数量： 为正数从头到尾找，为负数从尾到头找，都删除count个；为0移除所有。
		LRem(key string, count int, value string) (int, error)
		// LRemCtx 移除列表元素。count 控制移除的方向和数量： 为正数从头到尾找，为负数从尾到头找，都删除count个；为0移除所有。
		LRemCtx(ctx context.Context, key string, count int, value string) (val int, err error)
		// LTrim 修剪列表，只保留指定起止区间的元素。
		LTrim(key string, start, stop int64) error
		// LTrimCtx 修剪列表，只保留指定起止区间的元素。
		LTrimCtx(ctx context.Context, key string, start, stop int64) error
		// Persist 移除 key 的过期时间，key 将持久保持。
		Persist(key string) (bool, error)
		// PersistCtx 移除 key 的过期时间，key 将持久保持。
		PersistCtx(ctx context.Context, key string) (val bool, err error)
		// PFAdd 将 values 加入到键为 key 的 HyperLogLog 中，用于快速统计超大数据的唯一元素（基数）估算值。
		PFAdd(key string, values ...interface{}) (bool, error)
		// PFAddCtx 将 values 加入到键为 key 的 HyperLogLog 中，用于快速统计超大数据的唯一元素（基数）估算值。
		PFAddCtx(ctx context.Context, key string, values ...interface{}) (val bool, err error)
		// PFCount 返回给定 HyperLogLog 的基数估算值。
		PFCount(key string) (int64, error)
		// PFCountCtx 返回给定 HyperLogLog 的基数估算值。
		PFCountCtx(ctx context.Context, key string) (val int64, err error)
		// RPop 移除并返回列表的最后一个元素。
		RPop(key string) (string, error)
		// RPopCtx 移除并返回列表的最后一个元素。
		RPopCtx(ctx context.Context, key string) (val string, err error)
		// RPush 在列表右侧添加一个或多个值。
		RPush(key string, values ...interface{}) (int, error)
		// RPushCtx 在列表右侧添加一个或多个值。
		RPushCtx(ctx context.Context, key string, values ...interface{}) (val int, err error)
		// SAdd 向集合添加一个或多个成员。
		SAdd(key string, values ...interface{}) (int, error)
		// SAddCtx 向集合添加一个或多个成员。
		SAddCtx(ctx context.Context, key string, values ...interface{}) (val int, err error)
		// SScan 迭代集合中的元素。
		SScan(key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error)
		// SScanCtx 迭代集合中的元素。
		SScanCtx(ctx context.Context, key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error)
		// SCard 获取集合的成员数。
		SCard(key string) (int64, error)
		// SCardCtx 获取集合的成员数。
		SCardCtx(ctx context.Context, key string) (val int64, err error)
		// Set 设置 key 的值。
		Set(key, value string) error
		// SetCtx 设置 key 的值。
		SetCtx(ctx context.Context, key, value string) error
		// SetBit 设置或清除 key 上偏移量为 offset 的比特值为 value。
		SetBit(key string, offset int64, value int) (int, error)
		// SetBitCtx 设置或清除 key 上偏移量为 offset 的比特值为 value。
		SetBitCtx(ctx context.Context, key string, offset int64, value int) (int, error)
		// SetEx 设置键值及其存活秒数，过期会自动删除。
		SetEx(key, value string, seconds int) error
		// SetExCtx 设置键值及其存活秒数，过期会自动删除。
		SetExCtx(ctx context.Context, key, value string, seconds int) error
		// SetNX 当 key 不存在时，设置键值对。
		SetNX(key, value string) (bool, error)
		// SetNXCtx 当 key 不存在时，设置键值对。
		SetNXCtx(ctx context.Context, key, value string) (val bool, err error)
		// SetNXEx 当 key 不存时，设置键值对及其存活秒数，过期会自动删除。
		SetNXEx(key, value string, seconds int) (bool, error)
		// SetNXExCtx 当 key 不存时，设置键值对及其存活秒数，过期会自动删除。
		SetNXExCtx(ctx context.Context, key, value string, seconds int) (val bool, err error)
		// SIsMember 判断 member 是否为集合 key 的成员。
		SIsMember(key string, member interface{}) (bool, error)
		// SIsMemberCtx 判断 member 是否为集合 key 的成员。
		SIsMemberCtx(ctx context.Context, key string, member interface{}) (val bool, err error)
		// SMembers 返回集合 key 中的所有成员。
		SMembers(key string) ([]string, error)
		// SMembersCtx 返回集合 key 中的所有成员。
		SMembersCtx(ctx context.Context, key string) (val []string, err error)
		// SPop 随机移除并返回集合 key 中的一个成员。
		SPop(key string) (string, error)
		// SPopCtx 随机移除并返回集合 key 中的一个成员。
		SPopCtx(ctx context.Context, key string) (val string, err error)
		// SRandMember 随机返回集合 key 中的 count 个成员。
		SRandMember(key string, count int) ([]string, error)
		// SRandMemberCtx 随机返回集合 key 中的 count 个成员。
		SRandMemberCtx(ctx context.Context, key string, count int) (val []string, err error)
		// SRem 移除集合 key 中的一个或多个成员 members。
		SRem(key string, members ...interface{}) (int, error)
		// SRemCtx 移除集合 key 中的一个或多个成员 members。
		SRemCtx(ctx context.Context, key string, members ...interface{}) (val int, err error)
		// TTL 返回 key 的剩余生存秒数。
		TTL(key string) (int, error)
		// TTLCtx 返回 key 的剩余生存秒数。
		TTLCtx(ctx context.Context, key string) (val int, err error)
		// ZAdd 向有序集合 key 添加或更新一个成员及其分数。
		ZAdd(key string, score int64, member string) (bool, error)
		// ZAddFloat 向有序集合 key 添加或更新一个成员及其分数。
		ZAddFloat(key string, score float64, member string) (bool, error)
		// ZAddCtx 向有序集合 key 添加或更新一个成员及其分数。
		ZAddCtx(ctx context.Context, key string, score int64, member string) (val bool, err error)
		// ZAddFloatCtx 向有序集合 key 添加或更新一个成员及其分数。
		ZAddFloatCtx(ctx context.Context, key string, score float64, member string) (val bool, err error)
		// ZAdds 向有序集合添加或更新一个或多个成员及其分数。
		ZAdds(key string, ps ...redis.Pair) (int64, error)
		// ZAddsCtx 向有序集合添加或更新一个或多个成员及其分数。
		ZAddsCtx(ctx context.Context, key string, ps ...redis.Pair) (val int64, err error)
		// ZCard 获取有序集合 key 的成员数量。
		ZCard(key string) (int, error)
		// ZCardCtx 获取有序集合 key 的成员数量。
		ZCardCtx(ctx context.Context, key string) (val int, err error)
		// ZCount 求有序集合中指定分数区间的成员数量。
		ZCount(key string, start, stop int64) (int, error)
		// ZCountCtx 求有序集合中指定分数区间的成员数量。
		ZCountCtx(ctx context.Context, key string, start, stop int64) (val int, err error)
		// ZIncrBy 在有序集合中指定成员的分数上增加 increment 分。
		ZIncrBy(key string, increment int64, member string) (int64, error)
		// ZIncrByCtx 在有序集合中指定成员的分数上增加 increment 分。
		ZIncrByCtx(ctx context.Context, key string, increment int64, member string) (val int64, err error)
		// ZScore 返回有序集合中指定成员的分数。
		ZScore(key, value string) (int64, error)
		// ZScoreCtx 返回有序集合中指定成员的分数。
		ZScoreCtx(ctx context.Context, key, member string) (val int64, err error)
		// ZRank 获取有序集合中给定成员的升序索引排名。
		ZRank(key, field string) (int64, error)
		// ZRankCtx 获取有序集合中给定成员的升序索引排名。
		ZRankCtx(ctx context.Context, key, field string) (val int64, err error)
		// ZRem 移除有序集合中给定的一个或多个成员。
		ZRem(key string, members ...interface{}) (int, error)
		// ZRemCtx 移除有序集合中给定的一个或多个成员。
		ZRemCtx(ctx context.Context, key string, members ...interface{}) (val int, err error)
		// ZRemRangeByScore 移除有序集合中给定分数区间的所有成员。
		ZRemRangeByScore(key string, start, stop int64) (int, error)
		// ZRemRangeByScoreCtx 移除有序集合中给定分数区间的所有成员。
		ZRemRangeByScoreCtx(ctx context.Context, key string, start, stop int64) (val int, err error)
		// ZRemRangeByRank 移除有序集合中给定排名区间的所有成员。
		ZRemRangeByRank(key string, start, stop int64) (int, error)
		// ZRemRangeByRankCtx 移除有序集合中给定排名区间的所有成员。
		ZRemRangeByRankCtx(ctx context.Context, key string, start, stop int64) (val int, err error)
		// ZRange 获取有序集合中在索引区间内的所有成员。
		ZRange(key string, start, stop int64) ([]string, error)
		// ZRangeCtx 获取有序集合中在索引区间内的所有成员。
		ZRangeCtx(ctx context.Context, key string, start, stop int64) (val []string, err error)
		// ZRangeWithScores 获取有序集合中在索引区间内的所有成员及其分数。
		ZRangeWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZRangeWithScoresCtx 获取有序集合中在索引区间内的所有成员及其分数。
		ZRangeWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZRevRangeWithScores 获取有序集合中在索引区间内的所有成员及其分数，按索引从高到低排序。
		ZRevRangeWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZRevRangeWithScoresCtx 获取有序集合中在索引区间内的所有成员及其分数，按索引从高到低排序。
		ZRevRangeWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZRangeByScoreWithScores 获取有序集合中在分数区间内的所有成员及其分数。
		ZRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZRangeByScoreWithScoresCtx 获取有序集合中在分数区间内的所有成员及其分数。
		ZRangeByScoreWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZRangeByScoreWithScoresAndLimit 获取有序集合中在分数区间内的指定分页的成员及其分数。
		ZRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]redis.Pair, error)
		// ZRangeByScoreWithScoresAndLimitCtx 获取有序集合中在分数区间内的指定分页的成员及其分数。
		ZRangeByScoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (val []redis.Pair, err error)
		// ZRevRange 获取有序集合中在索引区间内的所有成员，按分数从高到低排序。
		ZRevRange(key string, start, stop int64) ([]string, error)
		// ZRevRangeCtx 获取有序集合中在索引区间内的所有成员，按分数从高到低排序。
		ZRevRangeCtx(ctx context.Context, key string, start, stop int64) (val []string, err error)
		// ZRevRangeByScoreWithScores 获取有序集合中在分数区间内的所有成员及其分数，按索引从高到低排序。
		ZRevRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZRevRangeByScoreWithScoresCtx 获取有序集合中在分数区间内的所有成员及其分数，按索引从高到低排序。
		ZRevRangeByScoreWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZRevRangeByScoreWithScoresAndLimit 获取有序集合中在分数区间内的指定分页的成员及其分数，按分数从高到低排序。
		ZRevRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]redis.Pair, error)
		// ZRevRangeByScoreWithScoresAndLimitCtx 获取有序集合中在分数区间内的指定分页的成员及其分数，按分数从高到低排序。
		ZRevRangeByScoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (val []redis.Pair, err error)
		// ZRevRank 获取有序集合中给定成员的降序索引排名。
		ZRevRank(key, member string) (int64, error)
		// ZRevRankCtx 获取有序集合中给定成员的降序索引排名。
		ZRevRankCtx(ctx context.Context, key, member string) (val int64, err error)
	}

	// 基于缓存集群的 kv 存储
	kvStore struct {
		dispatcher *hash.ConsistentHash
	}
)

// New 返回一个键值对存储 KvStore。
func New(c Config) Store {
	if len(c) == 0 || cache.TotalWeights(c) <= 0 {
		log.Fatal("未配置缓存节点")
	}

	// 即使只有一个节点，我们也是用一致性哈希，
	// 因为 kv存储和redis存储的方法不同。
	dispatcher := hash.NewConsistentHash()
	for _, cfg := range c {
		rds := cfg.NewRedis()
		dispatcher.AddWithWeight(rds, cfg.Weight)
	}

	return kvStore{
		dispatcher: dispatcher,
	}
}

func (s kvStore) Decr(key string) (int64, error) {
	return s.DecrCtx(context.Background(), key)
}

func (s kvStore) DecrCtx(ctx context.Context, key string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.DecrCtx(ctx, key)
}

func (s kvStore) DecrBy(key string, decrement int64) (int64, error) {
	return s.DecrByCtx(context.Background(), key, decrement)
}

func (s kvStore) DecrByCtx(ctx context.Context, key string, decrement int64) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.DecrByCtx(ctx, key, decrement)
}

func (s kvStore) Del(keys ...string) (int, error) {
	return s.DelCtx(context.Background(), keys...)
}

func (s kvStore) DelCtx(ctx context.Context, keys ...string) (int, error) {
	var val int
	var be errorx.BatchError

	for _, key := range keys {
		node, e := s.getRedis(key)
		if e != nil {
			be.Add(e)
			continue
		}

		if v, e := node.DelCtx(ctx, key); e != nil {
			be.Add(e)
		} else {
			val += v
		}
	}

	return val, be.Err()
}

func (s kvStore) Eval(script, key string, args ...interface{}) (interface{}, error) {
	return s.EvalCtx(context.Background(), script, key, args...)
}

func (s kvStore) EvalCtx(ctx context.Context, script, key string, args ...interface{}) (interface{}, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.EvalCtx(ctx, script, []string{key}, args...)
}

func (s kvStore) Exists(key string) (bool, error) {
	return s.ExistsCtx(context.Background(), key)
}

func (s kvStore) ExistsCtx(ctx context.Context, key string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.ExistsCtx(ctx, key)
}

func (s kvStore) Expire(key string, seconds int) error {
	return s.ExpireCtx(context.Background(), key, seconds)
}

func (s kvStore) ExpireCtx(ctx context.Context, key string, seconds int) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.ExpireCtx(ctx, key, seconds)
}

func (s kvStore) ExpireAt(key string, expireTime int64) error {
	return s.ExpireAtCtx(context.Background(), key, expireTime)
}

func (s kvStore) ExpireAtCtx(ctx context.Context, key string, expireTime int64) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.ExpireAtCtx(ctx, key, expireTime)
}

func (s kvStore) Get(key string) (string, error) {
	return s.GetCtx(context.Background(), key)
}

func (s kvStore) GetCtx(ctx context.Context, key string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.GetCtx(ctx, key)
}

func (s kvStore) HDel(key, field string) (bool, error) {
	return s.HDelCtx(context.Background(), key, field)
}

func (s kvStore) HDelCtx(ctx context.Context, key, field string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HDelCtx(ctx, key, field)
}

func (s kvStore) HExists(key, field string) (bool, error) {
	return s.HExistsCtx(context.Background(), key, field)
}

func (s kvStore) HExistsCtx(ctx context.Context, key, field string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HExistsCtx(ctx, key, field)
}

func (s kvStore) HGet(key, field string) (string, error) {
	return s.HGetCtx(context.Background(), key, field)
}

func (s kvStore) HGetCtx(ctx context.Context, key, field string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.HGetCtx(ctx, key, field)
}

func (s kvStore) HGetAll(key string) (map[string]string, error) {
	return s.HGetAllCtx(context.Background(), key)
}

func (s kvStore) HGetAllCtx(ctx context.Context, key string) (map[string]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HGetAllCtx(ctx, key)
}

func (s kvStore) HIncrBy(key, field string, increment int) (int, error) {
	return s.HIncrByCtx(context.Background(), key, field, increment)
}

func (s kvStore) HIncrByCtx(ctx context.Context, key, field string, increment int) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.HIncrByCtx(ctx, key, field, increment)
}

func (s kvStore) HKeys(key string) ([]string, error) {
	return s.HKeysCtx(context.Background(), key)
}

func (s kvStore) HKeysCtx(ctx context.Context, key string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HKeysCtx(ctx, key)
}

func (s kvStore) HLen(key string) (int, error) {
	return s.HLenCtx(context.Background(), key)
}

func (s kvStore) HLenCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.HLenCtx(ctx, key)
}

func (s kvStore) HMGet(key string, fields ...string) ([]string, error) {
	return s.HMGetCtx(context.Background(), key, fields...)
}

func (s kvStore) HMGetCtx(ctx context.Context, key string, fields ...string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HMGetCtx(ctx, key, fields...)
}

func (s kvStore) HSet(key, field, value string) error {
	return s.HSetCtx(context.Background(), key, field, value)
}

func (s kvStore) HSetCtx(ctx context.Context, key, field, value string) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.HSetCtx(ctx, key, field, value)
}

func (s kvStore) HSetNx(key, field, value string) (bool, error) {
	return s.HSetNxCtx(context.Background(), key, field, value)
}

func (s kvStore) HSetNxCtx(ctx context.Context, key, field, value string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HSetNXCtx(ctx, key, field, value)
}

func (s kvStore) HMSet(key string, fieldsAndValues map[string]string) error {
	return s.HMSetCtx(context.Background(), key, fieldsAndValues)
}

func (s kvStore) HMSetCtx(ctx context.Context, key string, fieldsAndValues map[string]string) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.HMSetCtx(ctx, key, fieldsAndValues)
}

func (s kvStore) HVals(key string) ([]string, error) {
	return s.HValsCtx(context.Background(), key)
}

func (s kvStore) HValsCtx(ctx context.Context, key string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HValsCtx(ctx, key)
}

func (s kvStore) Incr(key string) (int64, error) {
	return s.IncrCtx(context.Background(), key)
}

func (s kvStore) IncrCtx(ctx context.Context, key string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.IncrCtx(ctx, key)
}

func (s kvStore) IncrBy(key string, increment int64) (int64, error) {
	return s.IncrByCtx(context.Background(), key, increment)
}

func (s kvStore) IncrByCtx(ctx context.Context, key string, increment int64) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.IncrByCtx(ctx, key, increment)
}

func (s kvStore) LLen(key string) (int, error) {
	return s.LLenCtx(context.Background(), key)
}

func (s kvStore) LLenCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LLenCtx(ctx, key)
}

func (s kvStore) LIndex(key string, index int64) (string, error) {
	return s.LIndexCtx(context.Background(), key, index)
}

func (s kvStore) LIndexCtx(ctx context.Context, key string, index int64) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.LIndexCtx(ctx, key, index)
}

func (s kvStore) LPop(key string) (string, error) {
	return s.LPopCtx(context.Background(), key)
}

func (s kvStore) LPopCtx(ctx context.Context, key string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.LPopCtx(ctx, key)
}

func (s kvStore) LPush(key string, values ...interface{}) (int, error) {
	return s.LPushCtx(context.Background(), key, values...)
}

func (s kvStore) LPushCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LPushCtx(ctx, key, values...)
}

func (s kvStore) LRange(key string, start, stop int) ([]string, error) {
	return s.LRangeCtx(context.Background(), key, start, stop)
}

func (s kvStore) LRangeCtx(ctx context.Context, key string, start, stop int) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.LRangeCtx(ctx, key, start, stop)
}

func (s kvStore) LRem(key string, count int, value string) (int, error) {
	return s.LRemCtx(context.Background(), key, count, value)
}

func (s kvStore) LRemCtx(ctx context.Context, key string, count int, value string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LRemCtx(ctx, key, count, value)
}

func (s kvStore) LTrim(key string, start, stop int64) error {
	return s.LTrimCtx(context.Background(), key, start, stop)
}

func (s kvStore) LTrimCtx(ctx context.Context, key string, start, stop int64) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.LTrimCtx(ctx, key, start, stop)
}

func (s kvStore) Persist(key string) (bool, error) {
	return s.PersistCtx(context.Background(), key)
}

func (s kvStore) PersistCtx(ctx context.Context, key string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.PersistCtx(ctx, key)
}

func (s kvStore) PFAdd(key string, values ...interface{}) (bool, error) {
	return s.PFAddCtx(context.Background(), key, values...)
}

func (s kvStore) PFAddCtx(ctx context.Context, key string, values ...interface{}) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.PFAddCtx(ctx, key, values...)
}

func (s kvStore) PFCount(key string) (int64, error) {
	return s.PFCountCtx(context.Background(), key)
}

func (s kvStore) PFCountCtx(ctx context.Context, key string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.PFCountCtx(ctx, key)
}

func (s kvStore) RPop(key string) (string, error) {
	return s.RPopCtx(context.Background(), key)

}

func (s kvStore) RPopCtx(ctx context.Context, key string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.RPopCtx(ctx, key)
}

func (s kvStore) RPush(key string, values ...interface{}) (int, error) {
	return s.RPushCtx(context.Background(), key, values...)
}

func (s kvStore) RPushCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.RPushCtx(ctx, key, values...)
}

func (s kvStore) SAdd(key string, values ...interface{}) (int, error) {
	return s.SAddCtx(context.Background(), key, values...)
}

func (s kvStore) SAddCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SAddCtx(ctx, key, values...)
}

func (s kvStore) SCard(key string) (int64, error) {
	return s.SCardCtx(context.Background(), key)
}

func (s kvStore) SCardCtx(ctx context.Context, key string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SCardCtx(ctx, key)
}

func (s kvStore) Set(key, value string) error {
	return s.SetCtx(context.Background(), key, value)
}

func (s kvStore) SetCtx(ctx context.Context, key, value string) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.SetCtx(ctx, key, value)
}

func (s kvStore) SetBit(key string, offset int64, value int) (int, error) {
	return s.SetBitCtx(context.Background(), key, offset, value)
}

func (s kvStore) SetBitCtx(ctx context.Context, key string, offset int64, value int) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SetBitCtx(ctx, key, offset, value)
}

func (s kvStore) SetEx(key, value string, seconds int) error {
	return s.SetExCtx(context.Background(), key, value, seconds)
}

func (s kvStore) SetExCtx(ctx context.Context, key, value string, seconds int) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.SetExCtx(ctx, key, value, seconds)
}

func (s kvStore) SetNX(key, value string) (bool, error) {
	return s.SetNXCtx(context.Background(), key, value)
}

func (s kvStore) SetNXCtx(ctx context.Context, key, value string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SetNXCtx(ctx, key, value)
}

func (s kvStore) SetNXEx(key, value string, seconds int) (bool, error) {
	return s.SetNXExCtx(context.Background(), key, value, seconds)
}

func (s kvStore) SetNXExCtx(ctx context.Context, key, value string, seconds int) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SetNXExCtx(ctx, key, value, seconds)
}

func (s kvStore) GetSet(key, value string) (string, error) {
	return s.GetSetCtx(context.Background(), key, value)
}

func (s kvStore) GetSetCtx(ctx context.Context, key, value string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.GetSetCtx(ctx, key, value)
}

func (s kvStore) GetBit(key string, offset int64) (int, error) {
	return s.GetBitCtx(context.Background(), key, offset)
}

func (s kvStore) GetBitCtx(ctx context.Context, key string, offset int64) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.GetBitCtx(ctx, key, offset)
}

func (s kvStore) SIsMember(key string, value interface{}) (bool, error) {
	return s.SIsMemberCtx(context.Background(), key, value)
}

func (s kvStore) SIsMemberCtx(ctx context.Context, key string, value interface{}) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SIsMemberCtx(ctx, key, value)
}

func (s kvStore) SMembers(key string) ([]string, error) {
	return s.SMembersCtx(context.Background(), key)
}

func (s kvStore) SMembersCtx(ctx context.Context, key string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.SMembersCtx(ctx, key)
}

func (s kvStore) SPop(key string) (string, error) {
	return s.SPopCtx(context.Background(), key)
}

func (s kvStore) SPopCtx(ctx context.Context, key string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.SPopCtx(ctx, key)
}

func (s kvStore) SRandMember(key string, count int) ([]string, error) {
	return s.SRandMemberCtx(context.Background(), key, count)
}

func (s kvStore) SRandMemberCtx(ctx context.Context, key string, count int) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.SRandMemberCtx(ctx, key, count)
}

func (s kvStore) SRem(key string, values ...interface{}) (int, error) {
	return s.SRemCtx(context.Background(), key, values...)
}

func (s kvStore) SRemCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SRemCtx(ctx, key, values...)
}

func (s kvStore) SScan(key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	return s.SScanCtx(context.Background(), key, cursor, match, count)
}

func (s kvStore) SScanCtx(ctx context.Context, key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, 0, err
	}

	return node.SScanCtx(ctx, key, cursor, match, count)
}

func (s kvStore) TTL(key string) (int, error) {
	return s.TTLCtx(context.Background(), key)
}

func (s kvStore) TTLCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.TTLCtx(ctx, key)
}

func (s kvStore) ZAdd(key string, score int64, value string) (bool, error) {
	return s.ZAddCtx(context.Background(), key, score, value)
}

func (s kvStore) ZAddFloat(key string, score float64, value string) (bool, error) {
	return s.ZAddFloatCtx(context.Background(), key, score, value)
}

func (s kvStore) ZAddCtx(ctx context.Context, key string, score int64, value string) (bool, error) {
	return s.ZAddFloatCtx(ctx, key, float64(score), value)
}

func (s kvStore) ZAddFloatCtx(ctx context.Context, key string, score float64, value string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.ZAddFloatCtx(ctx, key, score, value)
}

func (s kvStore) ZAdds(key string, ps ...redis.Pair) (int64, error) {
	return s.ZAddsCtx(context.Background(), key, ps...)
}

func (s kvStore) ZAddsCtx(ctx context.Context, key string, ps ...redis.Pair) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZAddsCtx(ctx, key, ps...)
}

func (s kvStore) ZCard(key string) (int, error) {
	return s.ZCardCtx(context.Background(), key)
}

func (s kvStore) ZCardCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZCardCtx(ctx, key)
}

func (s kvStore) ZCount(key string, start, stop int64) (int, error) {
	return s.ZCountCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZCountCtx(ctx context.Context, key string, start, stop int64) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZCountCtx(ctx, key, start, stop)
}

func (s kvStore) ZIncrBy(key string, increment int64, field string) (int64, error) {
	return s.ZIncrByCtx(context.Background(), key, increment, field)
}

func (s kvStore) ZIncrByCtx(ctx context.Context, key string, increment int64, field string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZIncrByCtx(ctx, key, increment, field)
}

func (s kvStore) ZRank(key, field string) (int64, error) {
	return s.ZRankCtx(context.Background(), key, field)
}

func (s kvStore) ZRankCtx(ctx context.Context, key, field string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRankCtx(ctx, key, field)
}

func (s kvStore) ZRange(key string, start, stop int64) ([]string, error) {
	return s.ZRangeCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRangeCtx(ctx context.Context, key string, start, stop int64) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRangeCtx(ctx, key, start, stop)
}

func (s kvStore) ZRangeWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	return s.ZRangeWithScoresCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRangeWithScoresCtx(ctx context.Context, key string, start, stop int64) ([]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRangeWithScoresCtx(ctx, key, start, stop)
}

func (s kvStore) ZRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	return s.ZRangeByScoreWithScoresCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRangeByScoreWithScoresCtx(ctx context.Context, key string, start, stop int64) ([]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRangeByScoreWithScoresCtx(ctx, key, start, stop)
}

func (s kvStore) ZRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	return s.ZRangeByScoreWithScoresAndLimitCtx(context.Background(), key, start, stop, page, size)
}

func (s kvStore) ZRangeByScoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRangeByScoreWithScoresAndLimitCtx(ctx, key, start, stop, page, size)
}

func (s kvStore) ZRem(key string, values ...interface{}) (int, error) {
	return s.ZRemCtx(context.Background(), key, values...)
}

func (s kvStore) ZRemCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRemCtx(ctx, key, values...)
}

func (s kvStore) ZRemRangeByRank(key string, start, stop int64) (int, error) {
	return s.ZRemRangeByRankCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRemRangeByRankCtx(ctx context.Context, key string, start, stop int64) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRemRangeByRankCtx(ctx, key, start, stop)
}

func (s kvStore) ZRemRangeByScore(key string, start, stop int64) (int, error) {
	return s.ZRemRangeByScoreCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRemRangeByScoreCtx(ctx context.Context, key string, start, stop int64) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRemRangeByScoreCtx(ctx, key, start, stop)
}

func (s kvStore) ZRevRange(key string, start, stop int64) ([]string, error) {
	return s.ZRevRangeCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRevRangeCtx(ctx context.Context, key string, start, stop int64) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRevRangeCtx(ctx, key, start, stop)
}

func (s kvStore) ZRevRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	return s.ZRevRangeByScoreWithScoresCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRevRangeByScoreWithScoresCtx(ctx context.Context, key string, start, stop int64) ([]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRevRangeByScoreWithScoresCtx(ctx, key, start, stop)
}

func (s kvStore) ZRevRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	return s.ZRevRangeByScoreWithScoresAndLimitCtx(context.Background(), key, start, stop, page, size)
}

func (s kvStore) ZRevRangeByScoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRevRangeByScoreWithScoresAndLimitCtx(ctx, key, start, stop, page, size)
}

func (s kvStore) ZRevRank(key, field string) (int64, error) {
	return s.ZRevRankCtx(context.Background(), key, field)
}

func (s kvStore) ZRevRankCtx(ctx context.Context, key, field string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRevRankCtx(ctx, key, field)
}

func (s kvStore) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	return s.ZRevRangeWithScoresCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZRevRangeWithScoresCtx(ctx context.Context, key string, start, stop int64) ([]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRevRangeWithScoresCtx(ctx, key, start, stop)
}

func (s kvStore) ZScore(key, value string) (int64, error) {
	return s.ZScoreCtx(context.Background(), key, value)
}

func (s kvStore) ZScoreCtx(ctx context.Context, key, value string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZScoreCtx(ctx, key, value)
}

func (s kvStore) getRedis(key string) (*redis.Redis, error) {
	node, ok := s.dispatcher.Get(key)
	if !ok {
		return nil, ErrNoRedisNode
	}

	return node.(*redis.Redis), nil
}
