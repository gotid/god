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
		// Decrby 将 key 中存储的数值减去 decrement。
		Decrby(key string, decrement int64) (int64, error)
		// DecrbyCtx 将 key 中存储的数值减去 decrement。
		DecrbyCtx(ctx context.Context, key string, decrement int64) (int64, error)
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
		// Expireat 设置 key 的过期时间，过期会自动删除。
		Expireat(key string, expireTime int64) error
		// ExpireatCtx 设置 key 的过期时间，过期会自动删除。
		ExpireatCtx(ctx context.Context, key string, expireTime int64) error
		// Get 获取 key 的值。
		Get(key string) (string, error)
		// GetCtx 获取 key 的值。
		GetCtx(ctx context.Context, key string) (string, error)
		// GetSet 设置 key 的新值为 value，并返回就值。
		GetSet(key, value string) (string, error)
		// GetSetCtx 设置 key 的新值为 value，并返回就值。
		GetSetCtx(ctx context.Context, key, value string) (string, error)
		// Hdel 删除哈希 key 中的给定字段 fields。
		Hdel(key, field string) (bool, error)
		// HdelCtx 删除哈希 key 中的给定字段 fields。
		HdelCtx(ctx context.Context, key, field string) (bool, error)
		// Hexists 判断哈希 key 中成员 field 是否存在。
		Hexists(key, field string) (bool, error)
		// HexistsCtx 判断哈希 key 中成员 field 是否存在。
		HexistsCtx(ctx context.Context, key, field string) (bool, error)
		// Hget 获取哈希 key 中字段 field 的值。
		Hget(key, field string) (string, error)
		// HgetCtx 获取哈希 key 中字段 field 的值。
		HgetCtx(ctx context.Context, key, field string) (string, error)
		// Hgetall 获取哈希 key 的所有字段:值映射。
		Hgetall(key string) (map[string]string, error)
		// HgetallCtx 获取哈希 key 的所有字段:值映射。
		HgetallCtx(ctx context.Context, key string) (map[string]string, error)
		// Hincrby 为哈希 key 的字段 field 的值增加 increment。
		Hincrby(key, field string, increment int) (int, error)
		// HincrbyCtx 为哈希 key 的字段 field 的值增加 increment。
		HincrbyCtx(ctx context.Context, key, field string, increment int) (int, error)
		// Hkeys 返回哈希 key 的所有字段。
		Hkeys(key string) ([]string, error)
		// HkeysCtx 返回哈希 key 的所有字段。
		HkeysCtx(ctx context.Context, key string) ([]string, error)
		// Hlen 返回哈希 key 的字段数量。
		Hlen(key string) (int, error)
		// HlenCtx 返回哈希 key 的字段数量。
		HlenCtx(ctx context.Context, key string) (int, error)
		// Hmget 获取哈希 key 中所有给定字段 fields 的值。
		Hmget(key string, fields ...string) ([]string, error)
		// HmgetCtx 获取哈希 key 中所有给定字段 fields 的值。
		HmgetCtx(ctx context.Context, key string, fields ...string) ([]string, error)
		// Hset 设置哈希 key 的字段 field 的值为 value。
		Hset(key, field, value string) error
		// HsetCtx 设置哈希 key 的字段 field 的值为 value。
		HsetCtx(ctx context.Context, key, field, value string) error
		// Hsetnx 当哈希 key 中字段 field 不存在时，增加字段值 field:value。
		Hsetnx(key, field, value string) (bool, error)
		// HsetnxCtx 当哈希 key 中字段 field 不存在时，增加字段值 field:value。
		HsetnxCtx(ctx context.Context, key, field, value string) (bool, error)
		// Hmset 同时设置多个键值到哈希 key。
		Hmset(key string, fieldsAndValues map[string]string) error
		// HmsetCtx 同时设置多个键值到哈希 key。
		HmsetCtx(ctx context.Context, key string, fieldsAndValues map[string]string) error
		// Hvals 获取哈希表中所有值。
		Hvals(key string) ([]string, error)
		// HvalsCtx 获取哈希表中所有值。
		HvalsCtx(ctx context.Context, key string) ([]string, error)
		// Incr 将 key 中储存的数字值增一。
		Incr(key string) (int64, error)
		// IncrCtx 将 key 中储存的数字值增一。
		IncrCtx(ctx context.Context, key string) (int64, error)
		// Incrby 将 key 所储存的值加上给定的增量值（increment） 。
		Incrby(key string, increment int64) (int64, error)
		// IncrbyCtx 将 key 所储存的值加上给定的增量值（increment） 。
		IncrbyCtx(ctx context.Context, key string, increment int64) (int64, error)
		// Llen 获取列表长度。
		Llen(key string) (int, error)
		// LlenCtx 获取列表长度。
		LlenCtx(ctx context.Context, key string) (val int, err error)
		// Lindex 通过索引获取列表中的元素。
		Lindex(key string, index int64) (string, error)
		// LindexCtx 通过索引获取列表中的元素。
		LindexCtx(ctx context.Context, key string, index int64) (val string, err error)
		// Lpop 移除并返回列表的第一个元素。
		Lpop(key string) (string, error)
		// LpopCtx 移除并返回列表的第一个元素。
		LpopCtx(ctx context.Context, key string) (val string, err error)
		// Lpush 将一个或多个值插入到列表头部。
		Lpush(key string, values ...interface{}) (int, error)
		// LpushCtx 将一个或多个值插入到列表头部。
		LpushCtx(ctx context.Context, key string, values ...interface{}) (val int, err error)
		// Lrange 获取列表指定范围内的元素。
		Lrange(key string, start, stop int) ([]string, error)
		// LrangeCtx 获取列表指定范围内的元素。
		LrangeCtx(ctx context.Context, key string, start, stop int) (val []string, err error)
		// Lrem 移除列表元素。
		// count 控制移除的方向和数量： 为正数从头到尾找，为负数从尾到头找，都删除count个；为0移除所有。
		Lrem(key string, count int, value string) (int, error)
		// LremCtx 移除列表元素。
		// count 控制移除的方向和数量： 为正数从头到尾找，为负数从尾到头找，都删除count个；为0移除所有。
		LremCtx(ctx context.Context, key string, count int, value string) (val int, err error)
		// Ltrim 修剪列表，只保留指定起止区间的元素。
		Ltrim(key string, start, stop int64) error
		// LtrimCtx 修剪列表，只保留指定起止区间的元素。
		LtrimCtx(ctx context.Context, key string, start, stop int64) error
		// Persist 移除 key 的过期时间，key 将持久保持。
		Persist(key string) (bool, error)
		// PersistCtx 移除 key 的过期时间，key 将持久保持。
		PersistCtx(ctx context.Context, key string) (val bool, err error)
		// Pfadd 将 values 加入到键为 key 的 HyperLogLog 中，用于快速统计超大数据的唯一元素（基数）估算值。
		Pfadd(key string, values ...interface{}) (bool, error)
		// PfaddCtx 将 values 加入到键为 key 的 HyperLogLog 中，用于快速统计超大数据的唯一元素（基数）估算值。
		PfaddCtx(ctx context.Context, key string, values ...interface{}) (val bool, err error)
		// Pfcount 返回给定 HyperLogLog 的基数估算值。
		Pfcount(key string) (int64, error)
		// PfcountCtx 返回给定 HyperLogLog 的基数估算值。
		PfcountCtx(ctx context.Context, key string) (val int64, err error)
		// Rpop 移除并返回列表的最后一个元素。
		Rpop(key string) (string, error)
		// RpopCtx 移除并返回列表的最后一个元素。
		RpopCtx(ctx context.Context, key string) (val string, err error)
		// Rpush 在列表右侧添加一个或多个值。
		Rpush(key string, values ...interface{}) (int, error)
		// RpushCtx 在列表右侧添加一个或多个值。
		RpushCtx(ctx context.Context, key string, values ...interface{}) (val int, err error)
		// Sadd 向集合添加一个或多个成员。
		Sadd(key string, values ...interface{}) (int, error)
		// SaddCtx 向集合添加一个或多个成员。
		SaddCtx(ctx context.Context, key string, values ...interface{}) (val int, err error)
		// Sscan 迭代集合中的元素。
		Sscan(key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error)
		// SscanCtx 迭代集合中的元素。
		SscanCtx(ctx context.Context, key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error)
		// Scard 获取集合的成员数。
		Scard(key string) (int64, error)
		// ScardCtx 获取集合的成员数。
		ScardCtx(ctx context.Context, key string) (val int64, err error)
		// Set 设置 key 的值。
		Set(key, value string) error
		// SetCtx 设置 key 的值。
		SetCtx(ctx context.Context, key, value string) error
		// Setex 设置键值及其存活秒数，过期会自动删除。
		Setex(key, value string, seconds int) error
		// SetexCtx 设置键值及其存活秒数，过期会自动删除。
		SetexCtx(ctx context.Context, key, value string, seconds int) error
		// Setnx 当 key 不存在时，设置键值对。
		Setnx(key, value string) (bool, error)
		// SetnxCtx 当 key 不存在时，设置键值对。
		SetnxCtx(ctx context.Context, key, value string) (val bool, err error)
		// SetnxEx 当 key 不存时，设置键值对及其存活秒数，过期会自动删除。
		SetnxEx(key, value string, seconds int) (bool, error)
		// SetnxExCtx 当 key 不存时，设置键值对及其存活秒数，过期会自动删除。
		SetnxExCtx(ctx context.Context, key, value string, seconds int) (val bool, err error)
		// Sismember 判断 member 是否为集合 key 的成员。
		Sismember(key string, member interface{}) (bool, error)
		// SismemberCtx 判断 member 是否为集合 key 的成员。
		SismemberCtx(ctx context.Context, key string, member interface{}) (val bool, err error)
		// Smembers 返回集合 key 中的所有成员。
		Smembers(key string) ([]string, error)
		// SmembersCtx 返回集合 key 中的所有成员。
		SmembersCtx(ctx context.Context, key string) (val []string, err error)
		// Spop 随机移除并返回集合 key 中的一个成员。
		Spop(key string) (string, error)
		// SpopCtx 随机移除并返回集合 key 中的一个成员。
		SpopCtx(ctx context.Context, key string) (val string, err error)
		// Srandmember 随机返回集合 key 中的 count 个成员。
		Srandmember(key string, count int) ([]string, error)
		// SrandmemberCtx 随机返回集合 key 中的 count 个成员。
		SrandmemberCtx(ctx context.Context, key string, count int) (val []string, err error)
		// Srem 移除集合 key 中的一个或多个成员 members。
		Srem(key string, members ...interface{}) (int, error)
		// SremCtx 移除集合 key 中的一个或多个成员 members。
		SremCtx(ctx context.Context, key string, members ...interface{}) (val int, err error)
		// Ttl 返回 key 的剩余生存秒数。
		Ttl(key string) (int, error)
		// TtlCtx 返回 key 的剩余生存秒数。
		TtlCtx(ctx context.Context, key string) (val int, err error)
		// Zadd 向有序集合 key 添加或更新一个成员及其分数。
		Zadd(key string, score int64, member string) (bool, error)
		// ZaddFloat 向有序集合 key 添加或更新一个成员及其分数。
		ZaddFloat(key string, score float64, member string) (bool, error)
		// ZaddCtx 向有序集合 key 添加或更新一个成员及其分数。
		ZaddCtx(ctx context.Context, key string, score int64, member string) (val bool, err error)
		// ZaddFloatCtx 向有序集合 key 添加或更新一个成员及其分数。
		ZaddFloatCtx(ctx context.Context, key string, score float64, member string) (val bool, err error)
		// Zadds 向有序集合添加或更新一个或多个成员及其分数。
		Zadds(key string, ps ...redis.Pair) (int64, error)
		// ZaddsCtx 向有序集合添加或更新一个或多个成员及其分数。
		ZaddsCtx(ctx context.Context, key string, ps ...redis.Pair) (val int64, err error)
		// Zcard 获取有序集合 key 的成员数量。
		Zcard(key string) (int, error)
		// ZcardCtx 获取有序集合 key 的成员数量。
		ZcardCtx(ctx context.Context, key string) (val int, err error)
		// Zcount 求有序集合中指定分数区间的成员数量。
		Zcount(key string, start, stop int64) (int, error)
		// ZcountCtx 求有序集合中指定分数区间的成员数量。
		ZcountCtx(ctx context.Context, key string, start, stop int64) (val int, err error)
		// Zincrby 在有序集合中指定成员的分数上增加 increment 分。
		Zincrby(key string, increment int64, member string) (int64, error)
		// ZincrbyCtx 在有序集合中指定成员的分数上增加 increment 分。
		ZincrbyCtx(ctx context.Context, key string, increment int64, member string) (val int64, err error)
		// Zscore 返回有序集合中指定成员的分数。
		Zscore(key, value string) (int64, error)
		// ZscoreCtx 返回有序集合中指定成员的分数。
		ZscoreCtx(ctx context.Context, key, member string) (val int64, err error)
		// Zrank 获取有序集合中给定成员的升序索引排名。
		Zrank(key, field string) (int64, error)
		// ZrankCtx 获取有序集合中给定成员的升序索引排名。
		ZrankCtx(ctx context.Context, key, field string) (val int64, err error)
		// Zrem 移除有序集合中给定的一个或多个成员。
		Zrem(key string, members ...interface{}) (int, error)
		// ZremCtx 移除有序集合中给定的一个或多个成员。
		ZremCtx(ctx context.Context, key string, members ...interface{}) (val int, err error)
		// Zremrangebyscore 移除有序集合中给定分数区间的所有成员。
		Zremrangebyscore(key string, start, stop int64) (int, error)
		// ZremrangebyscoreCtx 移除有序集合中给定分数区间的所有成员。
		ZremrangebyscoreCtx(ctx context.Context, key string, start, stop int64) (val int, err error)
		// Zremrangebyrank 移除有序集合中给定排名区间的所有成员。
		Zremrangebyrank(key string, start, stop int64) (int, error)
		// ZremrangebyrankCtx 移除有序集合中给定排名区间的所有成员。
		ZremrangebyrankCtx(ctx context.Context, key string, start, stop int64) (val int, err error)
		// Zrange 获取有序集合中在索引区间内的所有成员。
		Zrange(key string, start, stop int64) ([]string, error)
		// ZrangeCtx 获取有序集合中在索引区间内的所有成员。
		ZrangeCtx(ctx context.Context, key string, start, stop int64) (val []string, err error)
		// ZrangeWithScores 获取有序集合中在索引区间内的所有成员及其分数。
		ZrangeWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZrangeWithScoresCtx 获取有序集合中在索引区间内的所有成员及其分数。
		ZrangeWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZRevRangeWithScores 获取有序集合中在索引区间内的所有成员及其分数，按索引从高到低排序。
		ZRevRangeWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZRevRangeWithScoresCtx 获取有序集合中在索引区间内的所有成员及其分数，按索引从高到低排序。
		ZRevRangeWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZrangebyscoreWithScores 获取有序集合中在分数区间内的所有成员及其分数。
		ZrangebyscoreWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZrangebyscoreWithScoresCtx 获取有序集合中在分数区间内的所有成员及其分数。
		ZrangebyscoreWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZrangebyscoreWithScoresAndLimit 获取有序集合中在分数区间内的指定分页的成员及其分数。
		ZrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]redis.Pair, error)
		// ZrangebyscoreWithScoresAndLimitCtx 获取有序集合中在分数区间内的指定分页的成员及其分数。
		ZrangebyscoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (val []redis.Pair, err error)
		// Zrevrange 获取有序集合中在索引区间内的所有成员，按分数从高到低排序。
		Zrevrange(key string, start, stop int64) ([]string, error)
		// ZrevrangeCtx 获取有序集合中在索引区间内的所有成员，按分数从高到低排序。
		ZrevrangeCtx(ctx context.Context, key string, start, stop int64) (val []string, err error)
		// ZrevrangebyscoreWithScores 获取有序集合中在分数区间内的所有成员及其分数，按索引从高到低排序。
		ZrevrangebyscoreWithScores(key string, start, stop int64) ([]redis.Pair, error)
		// ZrevrangebyscoreWithScoresCtx 获取有序集合中在分数区间内的所有成员及其分数，按索引从高到低排序。
		ZrevrangebyscoreWithScoresCtx(ctx context.Context, key string, start, stop int64) (val []redis.Pair, err error)
		// ZrevrangebyscoreWithScoresAndLimit 获取有序集合中在分数区间内的指定分页的成员及其分数，按分数从高到低排序。
		ZrevrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]redis.Pair, error)
		// ZrevrangebyscoreWithScoresAndLimitCtx 获取有序集合中在分数区间内的指定分页的成员及其分数，按分数从高到低排序。
		ZrevrangebyscoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (val []redis.Pair, err error)
		// Zrevrank 获取有序集合中给定成员的降序索引排名。
		Zrevrank(key, member string) (int64, error)
		// ZrevrankCtx 获取有序集合中给定成员的降序索引排名。
		ZrevrankCtx(ctx context.Context, key, member string) (val int64, err error)
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

func (s kvStore) Decrby(key string, decrement int64) (int64, error) {
	return s.DecrbyCtx(context.Background(), key, decrement)
}

func (s kvStore) DecrbyCtx(ctx context.Context, key string, decrement int64) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.DecrbyCtx(ctx, key, decrement)
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

func (s kvStore) Expireat(key string, expireTime int64) error {
	return s.ExpireatCtx(context.Background(), key, expireTime)
}

func (s kvStore) ExpireatCtx(ctx context.Context, key string, expireTime int64) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.ExpireatCtx(ctx, key, expireTime)
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

func (s kvStore) Hdel(key, field string) (bool, error) {
	return s.HdelCtx(context.Background(), key, field)
}

func (s kvStore) HdelCtx(ctx context.Context, key, field string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HdelCtx(ctx, key, field)
}

func (s kvStore) Hexists(key, field string) (bool, error) {
	return s.HexistsCtx(context.Background(), key, field)
}

func (s kvStore) HexistsCtx(ctx context.Context, key, field string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HexistsCtx(ctx, key, field)
}

func (s kvStore) Hget(key, field string) (string, error) {
	return s.HgetCtx(context.Background(), key, field)
}

func (s kvStore) HgetCtx(ctx context.Context, key, field string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.HgetCtx(ctx, key, field)
}

func (s kvStore) Hgetall(key string) (map[string]string, error) {
	return s.HgetallCtx(context.Background(), key)
}

func (s kvStore) HgetallCtx(ctx context.Context, key string) (map[string]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HgetallCtx(ctx, key)
}

func (s kvStore) Hincrby(key, field string, increment int) (int, error) {
	return s.HincrbyCtx(context.Background(), key, field, increment)
}

func (s kvStore) HincrbyCtx(ctx context.Context, key, field string, increment int) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.HincrbyCtx(ctx, key, field, increment)
}

func (s kvStore) Hkeys(key string) ([]string, error) {
	return s.HkeysCtx(context.Background(), key)
}

func (s kvStore) HkeysCtx(ctx context.Context, key string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HkeysCtx(ctx, key)
}

func (s kvStore) Hlen(key string) (int, error) {
	return s.HlenCtx(context.Background(), key)
}

func (s kvStore) HlenCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.HlenCtx(ctx, key)
}

func (s kvStore) Hmget(key string, fields ...string) ([]string, error) {
	return s.HmgetCtx(context.Background(), key, fields...)
}

func (s kvStore) HmgetCtx(ctx context.Context, key string, fields ...string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HmgetCtx(ctx, key, fields...)
}

func (s kvStore) Hset(key, field, value string) error {
	return s.HsetCtx(context.Background(), key, field, value)
}

func (s kvStore) HsetCtx(ctx context.Context, key, field, value string) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.HsetCtx(ctx, key, field, value)
}

func (s kvStore) Hsetnx(key, field, value string) (bool, error) {
	return s.HsetnxCtx(context.Background(), key, field, value)
}

func (s kvStore) HsetnxCtx(ctx context.Context, key, field, value string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HsetnxCtx(ctx, key, field, value)
}

func (s kvStore) Hmset(key string, fieldsAndValues map[string]string) error {
	return s.HmsetCtx(context.Background(), key, fieldsAndValues)
}

func (s kvStore) HmsetCtx(ctx context.Context, key string, fieldsAndValues map[string]string) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.HmsetCtx(ctx, key, fieldsAndValues)
}

func (s kvStore) Hvals(key string) ([]string, error) {
	return s.HvalsCtx(context.Background(), key)
}

func (s kvStore) HvalsCtx(ctx context.Context, key string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HvalsCtx(ctx, key)
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

func (s kvStore) Incrby(key string, increment int64) (int64, error) {
	return s.IncrbyCtx(context.Background(), key, increment)
}

func (s kvStore) IncrbyCtx(ctx context.Context, key string, increment int64) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.IncrbyCtx(ctx, key, increment)
}

func (s kvStore) Llen(key string) (int, error) {
	return s.LlenCtx(context.Background(), key)
}

func (s kvStore) LlenCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LlenCtx(ctx, key)
}

func (s kvStore) Lindex(key string, index int64) (string, error) {
	return s.LindexCtx(context.Background(), key, index)
}

func (s kvStore) LindexCtx(ctx context.Context, key string, index int64) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.LindexCtx(ctx, key, index)
}

func (s kvStore) Lpop(key string) (string, error) {
	return s.LpopCtx(context.Background(), key)
}

func (s kvStore) LpopCtx(ctx context.Context, key string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.LpopCtx(ctx, key)
}

func (s kvStore) Lpush(key string, values ...interface{}) (int, error) {
	return s.LpushCtx(context.Background(), key, values...)
}

func (s kvStore) LpushCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LpushCtx(ctx, key, values...)
}

func (s kvStore) Lrange(key string, start, stop int) ([]string, error) {
	return s.LrangeCtx(context.Background(), key, start, stop)
}

func (s kvStore) LrangeCtx(ctx context.Context, key string, start, stop int) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.LrangeCtx(ctx, key, start, stop)
}

func (s kvStore) Lrem(key string, count int, value string) (int, error) {
	return s.LremCtx(context.Background(), key, count, value)
}

func (s kvStore) LremCtx(ctx context.Context, key string, count int, value string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LremCtx(ctx, key, count, value)
}

func (s kvStore) Ltrim(key string, start, stop int64) error {
	return s.LtrimCtx(context.Background(), key, start, stop)
}

func (s kvStore) LtrimCtx(ctx context.Context, key string, start, stop int64) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.LtrimCtx(ctx, key, start, stop)
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

func (s kvStore) Pfadd(key string, values ...interface{}) (bool, error) {
	return s.PfaddCtx(context.Background(), key, values...)
}

func (s kvStore) PfaddCtx(ctx context.Context, key string, values ...interface{}) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.PfaddCtx(ctx, key, values...)
}

func (s kvStore) Pfcount(key string) (int64, error) {
	return s.PfcountCtx(context.Background(), key)
}

func (s kvStore) PfcountCtx(ctx context.Context, key string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.PfcountCtx(ctx, key)
}

func (s kvStore) Rpop(key string) (string, error) {
	return s.RpopCtx(context.Background(), key)

}

func (s kvStore) RpopCtx(ctx context.Context, key string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.RpopCtx(ctx, key)
}

func (s kvStore) Rpush(key string, values ...interface{}) (int, error) {
	return s.RpushCtx(context.Background(), key, values...)
}

func (s kvStore) RpushCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.RpushCtx(ctx, key, values...)
}

func (s kvStore) Sadd(key string, values ...interface{}) (int, error) {
	return s.SaddCtx(context.Background(), key, values...)
}

func (s kvStore) SaddCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SaddCtx(ctx, key, values...)
}

func (s kvStore) Scard(key string) (int64, error) {
	return s.ScardCtx(context.Background(), key)
}

func (s kvStore) ScardCtx(ctx context.Context, key string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ScardCtx(ctx, key)
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

func (s kvStore) Setex(key, value string, seconds int) error {
	return s.SetexCtx(context.Background(), key, value, seconds)
}

func (s kvStore) SetexCtx(ctx context.Context, key, value string, seconds int) error {
	node, err := s.getRedis(key)
	if err != nil {
		return err
	}

	return node.SetexCtx(ctx, key, value, seconds)
}

func (s kvStore) Setnx(key, value string) (bool, error) {
	return s.SetnxCtx(context.Background(), key, value)
}

func (s kvStore) SetnxCtx(ctx context.Context, key, value string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SetnxCtx(ctx, key, value)
}

func (s kvStore) SetnxEx(key, value string, seconds int) (bool, error) {
	return s.SetnxExCtx(context.Background(), key, value, seconds)
}

func (s kvStore) SetnxExCtx(ctx context.Context, key, value string, seconds int) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SetnxExCtx(ctx, key, value, seconds)
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

func (s kvStore) Sismember(key string, value interface{}) (bool, error) {
	return s.SismemberCtx(context.Background(), key, value)
}

func (s kvStore) SismemberCtx(ctx context.Context, key string, value interface{}) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SismemberCtx(ctx, key, value)
}

func (s kvStore) Smembers(key string) ([]string, error) {
	return s.SmembersCtx(context.Background(), key)
}

func (s kvStore) SmembersCtx(ctx context.Context, key string) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.SmembersCtx(ctx, key)
}

func (s kvStore) Spop(key string) (string, error) {
	return s.SpopCtx(context.Background(), key)
}

func (s kvStore) SpopCtx(ctx context.Context, key string) (string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.SpopCtx(ctx, key)
}

func (s kvStore) Srandmember(key string, count int) ([]string, error) {
	return s.SrandmemberCtx(context.Background(), key, count)
}

func (s kvStore) SrandmemberCtx(ctx context.Context, key string, count int) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.SrandmemberCtx(ctx, key, count)
}

func (s kvStore) Srem(key string, values ...interface{}) (int, error) {
	return s.SremCtx(context.Background(), key, values...)
}

func (s kvStore) SremCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SremCtx(ctx, key, values...)
}

func (s kvStore) Sscan(key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	return s.SscanCtx(context.Background(), key, cursor, match, count)
}

func (s kvStore) SscanCtx(ctx context.Context, key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, 0, err
	}

	return node.SscanCtx(ctx, key, cursor, match, count)
}

func (s kvStore) Ttl(key string) (int, error) {
	return s.TtlCtx(context.Background(), key)
}

func (s kvStore) TtlCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.TtlCtx(ctx, key)
}

func (s kvStore) Zadd(key string, score int64, value string) (bool, error) {
	return s.ZaddCtx(context.Background(), key, score, value)
}

func (s kvStore) ZaddFloat(key string, score float64, value string) (bool, error) {
	return s.ZaddFloatCtx(context.Background(), key, score, value)
}

func (s kvStore) ZaddCtx(ctx context.Context, key string, score int64, value string) (bool, error) {
	return s.ZaddFloatCtx(ctx, key, float64(score), value)
}

func (s kvStore) ZaddFloatCtx(ctx context.Context, key string, score float64, value string) (bool, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.ZaddFloatCtx(ctx, key, score, value)
}

func (s kvStore) Zadds(key string, ps ...redis.Pair) (int64, error) {
	return s.ZaddsCtx(context.Background(), key, ps...)
}

func (s kvStore) ZaddsCtx(ctx context.Context, key string, ps ...redis.Pair) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZaddsCtx(ctx, key, ps...)
}

func (s kvStore) Zcard(key string) (int, error) {
	return s.ZcardCtx(context.Background(), key)
}

func (s kvStore) ZcardCtx(ctx context.Context, key string) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZcardCtx(ctx, key)
}

func (s kvStore) Zcount(key string, start, stop int64) (int, error) {
	return s.ZcountCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZcountCtx(ctx context.Context, key string, start, stop int64) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZcountCtx(ctx, key, start, stop)
}

func (s kvStore) Zincrby(key string, increment int64, field string) (int64, error) {
	return s.ZincrbyCtx(context.Background(), key, increment, field)
}

func (s kvStore) ZincrbyCtx(ctx context.Context, key string, increment int64, field string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZincrbyCtx(ctx, key, increment, field)
}

func (s kvStore) Zrank(key, field string) (int64, error) {
	return s.ZrankCtx(context.Background(), key, field)
}

func (s kvStore) ZrankCtx(ctx context.Context, key, field string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZrankCtx(ctx, key, field)
}

func (s kvStore) Zrange(key string, start, stop int64) ([]string, error) {
	return s.ZrangeCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZrangeCtx(ctx context.Context, key string, start, stop int64) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZrangeCtx(ctx, key, start, stop)
}

func (s kvStore) ZrangeWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	return s.ZrangeWithScoresCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZrangeWithScoresCtx(ctx context.Context, key string, start, stop int64) ([]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZrangeWithScoresCtx(ctx, key, start, stop)
}

func (s kvStore) ZrangebyscoreWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	return s.ZrangebyscoreWithScoresCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZrangebyscoreWithScoresCtx(ctx context.Context, key string, start, stop int64) ([]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZrangebyscoreWithScoresCtx(ctx, key, start, stop)
}

func (s kvStore) ZrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	return s.ZrangebyscoreWithScoresAndLimitCtx(context.Background(), key, start, stop, page, size)
}

func (s kvStore) ZrangebyscoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZrangebyscoreWithScoresAndLimitCtx(ctx, key, start, stop, page, size)
}

func (s kvStore) Zrem(key string, values ...interface{}) (int, error) {
	return s.ZremCtx(context.Background(), key, values...)
}

func (s kvStore) ZremCtx(ctx context.Context, key string, values ...interface{}) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZremCtx(ctx, key, values...)
}

func (s kvStore) Zremrangebyrank(key string, start, stop int64) (int, error) {
	return s.ZremrangebyrankCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZremrangebyrankCtx(ctx context.Context, key string, start, stop int64) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZremrangebyrankCtx(ctx, key, start, stop)
}

func (s kvStore) Zremrangebyscore(key string, start, stop int64) (int, error) {
	return s.ZremrangebyscoreCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZremrangebyscoreCtx(ctx context.Context, key string, start, stop int64) (int, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZremrangebyscoreCtx(ctx, key, start, stop)
}

func (s kvStore) Zrevrange(key string, start, stop int64) ([]string, error) {
	return s.ZrevrangeCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZrevrangeCtx(ctx context.Context, key string, start, stop int64) ([]string, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZrevrangeCtx(ctx, key, start, stop)
}

func (s kvStore) ZrevrangebyscoreWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	return s.ZrevrangebyscoreWithScoresCtx(context.Background(), key, start, stop)
}

func (s kvStore) ZrevrangebyscoreWithScoresCtx(ctx context.Context, key string, start, stop int64) ([]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZrevrangebyscoreWithScoresCtx(ctx, key, start, stop)
}

func (s kvStore) ZrevrangebyscoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	return s.ZrevrangebyscoreWithScoresAndLimitCtx(context.Background(), key, start, stop, page, size)
}

func (s kvStore) ZrevrangebyscoreWithScoresAndLimitCtx(ctx context.Context, key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, start, stop, page, size)
}

func (s kvStore) Zrevrank(key, field string) (int64, error) {
	return s.ZrevrankCtx(context.Background(), key, field)
}

func (s kvStore) ZrevrankCtx(ctx context.Context, key, field string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZrevrankCtx(ctx, key, field)
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

func (s kvStore) Zscore(key, value string) (int64, error) {
	return s.ZscoreCtx(context.Background(), key, value)
}

func (s kvStore) ZscoreCtx(ctx context.Context, key, value string) (int64, error) {
	node, err := s.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZscoreCtx(ctx, key, value)
}

func (s kvStore) getRedis(key string) (*redis.Redis, error) {
	node, ok := s.dispatcher.Get(key)
	if !ok {
		return nil, ErrNoRedisNode
	}

	return node.(*redis.Redis), nil
}
