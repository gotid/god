package kv

import (
	"errors"
	"log"

	"git.zc0901.com/go/god/lib/errorx"
	"git.zc0901.com/go/god/lib/hash"
	"git.zc0901.com/go/god/lib/store/cache"
	"git.zc0901.com/go/god/lib/store/redis"
)

var ErrNoRedisNode = errors.New("无可用 redis 节点")

type (
	Store interface {
		Del(keys ...string) (int, error)
		Eval(script string, key string, args ...interface{}) (interface{}, error)
		Exists(key string) (bool, error)
		Expire(key string, seconds int) error
		ExpireAt(key string, expireTime int64) error
		Get(key string) (string, error)
		GetBit(key string, offset int64) (result int, err error)
		MGet(keys ...string) ([]string, error)
		HDel(key, field string) (bool, error)
		HExists(key, field string) (bool, error)
		HGet(key, field string) (string, error)
		HGetAll(key string) (map[string]string, error)
		HIncrBy(key, field string, increment int) (int, error)
		HKeys(key string) ([]string, error)
		HLen(key string) (int, error)
		HMGet(key string, fields ...string) ([]string, error)
		HSet(key, field, value string) error
		HSetNX(key, field, value string) (bool, error)
		HMSet(key string, fieldsAndValues map[string]string) error
		HVals(key string) ([]string, error)
		Incr(key string) (int64, error)
		IncrBy(key string, increment int64) (int64, error)
		LLen(key string) (int, error)
		LPop(key string) (string, error)
		LPush(key string, values ...interface{}) (int, error)
		LRange(key string, start int, stop int) ([]string, error)
		LRem(key string, count int, value string) (int, error)
		Persist(key string) (bool, error)
		PFAdd(key string, values ...interface{}) (bool, error)
		PFCount(key string) (int64, error)
		RPush(key string, values ...interface{}) (int, error)
		SAdd(key string, values ...interface{}) (int, error)
		SCard(key string) (int64, error)
		Set(key string, value string) error
		SetBit(key string, offset int64, value int) error
		SetBits(key string, offset []int64) error
		SetEx(key, value string, seconds int) error
		SetNX(key, value string) (bool, error)
		SetNXEx(key, value string, seconds int) (bool, error)
		SIsMember(key string, value interface{}) (bool, error)
		SMembers(key string) ([]string, error)
		SPop(key string) (string, error)
		SRandMemberN(key string, count int) ([]string, error)
		SRem(key string, values ...interface{}) (int, error)
		SScan(key string, cursor uint64, match string, count int64) (keys []string, cur uint64, err error)
		TTL(key string) (int, error)
		ZAdd(key string, score int64, value string) (bool, error)
		ZAdds(key string, ps ...redis.Pair) (int64, error)
		ZCard(key string) (int, error)
		ZCount(key string, start, stop int64) (int, error)
		ZIncrBy(key string, increment int64, field string) (int64, error)
		ZRange(key string, start, stop int64) ([]string, error)
		ZRangeWithScores(key string, start, stop int64) ([]redis.Pair, error)
		ZRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error)
		ZRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]redis.Pair, error)
		ZRank(key, field string) (int64, error)
		ZRem(key string, values ...interface{}) (int, error)
		ZRemRangeByRank(key string, start, stop int64) (int, error)
		ZRemRangeByScore(key string, start, stop int64) (int, error)
		ZRevRange(key string, start, stop int64) ([]string, error)
		ZRevRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error)
		ZRevRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) ([]redis.Pair, error)
		ZScore(key string, value string) (int64, error)
		ZRevRank(key, field string) (int64, error)
	}

	clusterStore struct {
		dispatcher *hash.ConsistentHash
	}
)

func NewStore(c KvConf) Store {
	if len(c) == 0 || cache.TotalWeights(c) <= 0 {
		log.Fatal("无可用缓存节点")
	}

	// even if only one node, we chose to use consistent hash,
	// because Store and redis.Redis has different methods.
	dispatcher := hash.NewConsistentHash()
	for _, node := range c {
		cn := node.NewRedis()
		dispatcher.AddWithWeight(cn, node.Weight)
	}

	return clusterStore{
		dispatcher: dispatcher,
	}
}

func (cs clusterStore) Del(keys ...string) (int, error) {
	var val int
	var es errorx.Errors

	for _, key := range keys {
		node, e := cs.getRedis(key)
		if e != nil {
			es.Add(e)
			continue
		}

		if v, e := node.Del(key); e != nil {
			es.Add(e)
		} else {
			val += v
		}
	}

	return val, es.Error()
}

func (cs clusterStore) Eval(script string, key string, args ...interface{}) (interface{}, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.Eval(script, []string{key}, args...)
}

func (cs clusterStore) Exists(key string) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.Exists(key)
}

func (cs clusterStore) Expire(key string, seconds int) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.Expire(key, seconds)
}

func (cs clusterStore) ExpireAt(key string, expireTime int64) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.ExpireAt(key, expireTime)
}

func (cs clusterStore) Get(key string) (string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.Get(key)
}

func (cs clusterStore) GetBit(key string, offset int64) (result int, err error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.GetBit(key, offset)
}

func (cs clusterStore) GetBits(key string, offset []int64) (result map[int64]bool, err error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.GetBits(key, offset)
}

func (cs clusterStore) Del2(keys ...string) (int, error) {
	var val int
	var es errorx.Errors

	for _, key := range keys {
		node, e := cs.getRedis(key)
		if e != nil {
			es.Add(e)
			continue
		}

		if v, e := node.Del(key); e != nil {
			es.Add(e)
		} else {
			val += v
		}
	}

	return val, es.Error()
}

func (cs clusterStore) MGet(keys ...string) (ret []string, err error) {
	if len(keys) == 0 {
		return nil, nil
	}

	var es errorx.Errors

	for _, key := range keys {
		node, err := cs.getRedis(key)
		if err != nil {
			es.Add(err)
			continue
		}

		if v, err := node.Get(key); err != nil {
			es.Add(err)
			ret = append(ret, "")
		} else {
			ret = append(ret, v)
		}
	}

	err = es.Error()
	return
}

func (cs clusterStore) HDel(key, field string) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HDel(key, field)
}

func (cs clusterStore) HExists(key, field string) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HExists(key, field)
}

func (cs clusterStore) HGet(key, field string) (string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.HGet(key, field)
}

func (cs clusterStore) HGetAll(key string) (map[string]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HGetAll(key)
}

func (cs clusterStore) HIncrBy(key, field string, increment int) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.HIncrBy(key, field, increment)
}

func (cs clusterStore) HKeys(key string) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HKeys(key)
}

func (cs clusterStore) HLen(key string) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.HLen(key)
}

func (cs clusterStore) HMGet(key string, fields ...string) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HMGet(key, fields...)
}

func (cs clusterStore) HSet(key, field, value string) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.HSet(key, field, value)
}

func (cs clusterStore) HSetNX(key, field, value string) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.HSetNX(key, field, value)
}

func (cs clusterStore) HMSet(key string, fieldsAndValues map[string]string) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.HMSet(key, fieldsAndValues)
}

func (cs clusterStore) HVals(key string) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.HVals(key)
}

func (cs clusterStore) Incr(key string) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.Incr(key)
}

func (cs clusterStore) IncrBy(key string, increment int64) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.IncrBy(key, increment)
}

func (cs clusterStore) LLen(key string) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LLen(key)
}

func (cs clusterStore) LPop(key string) (string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.LPop(key)
}

func (cs clusterStore) LPush(key string, values ...interface{}) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LPush(key, values...)
}

func (cs clusterStore) LRange(key string, start int, stop int) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.LRange(key, start, stop)
}

func (cs clusterStore) LRem(key string, count int, value string) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.LRem(key, count, value)
}

func (cs clusterStore) Persist(key string) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.Persist(key)
}

func (cs clusterStore) PFAdd(key string, values ...interface{}) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.PFAdd(key, values...)
}

func (cs clusterStore) PFCount(key string) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.PFCount(key)
}

func (cs clusterStore) RPush(key string, values ...interface{}) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.RPush(key, values...)
}

func (cs clusterStore) SAdd(key string, values ...interface{}) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SAdd(key, values...)
}

func (cs clusterStore) SCard(key string) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SCard(key)
}

func (cs clusterStore) Set(key string, value string) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.Set(key, value)
}

func (cs clusterStore) SetBit(key string, offset int64, value int) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.SetBit(key, offset, value)
}

func (cs clusterStore) SetBits(key string, offset []int64) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.SetBits(key, offset)
}

func (cs clusterStore) SetEx(key, value string, seconds int) error {
	node, err := cs.getRedis(key)
	if err != nil {
		return err
	}

	return node.SetEx(key, value, seconds)
}

func (cs clusterStore) SetNX(key, value string) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SetNX(key, value)
}

func (cs clusterStore) SetNXEx(key, value string, seconds int) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SetNXEx(key, value, seconds)
}

func (cs clusterStore) SIsMember(key string, value interface{}) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.SIsMember(key, value)
}

func (cs clusterStore) SMembers(key string) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.SMembers(key)
}

func (cs clusterStore) SPop(key string) (string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return "", err
	}

	return node.SPop(key)
}

func (cs clusterStore) SRandMemberN(key string, count int) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.SRandMemberN(key, count)
}

func (cs clusterStore) SRem(key string, values ...interface{}) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.SRem(key, values...)
}

func (cs clusterStore) SScan(key string, cursor uint64, match string, count int64) (
	keys []string, cur uint64, err error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, 0, err
	}

	return node.SScan(key, cursor, match, count)
}

func (cs clusterStore) TTL(key string) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.TTL(key)
}

func (cs clusterStore) ZAdd(key string, score int64, value string) (bool, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return false, err
	}

	return node.ZAdd(key, score, value)
}

func (cs clusterStore) ZAdds(key string, ps ...redis.Pair) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZAdds(key, ps...)
}

func (cs clusterStore) ZCard(key string) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZCard(key)
}

func (cs clusterStore) ZCount(key string, start, stop int64) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZCount(key, start, stop)
}

func (cs clusterStore) ZIncrBy(key string, increment int64, field string) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZIncrBy(key, increment, field)
}

func (cs clusterStore) ZRank(key, field string) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRank(key, field)
}

func (cs clusterStore) ZRange(key string, start, stop int64) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRange(key, start, stop)
}

func (cs clusterStore) ZRangeWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRangeWithScores(key, start, stop)
}

func (cs clusterStore) ZRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRangeByScoreWithScores(key, start, stop)
}

func (cs clusterStore) ZRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRangeByScoreWithScoresAndLimit(key, start, stop, page, size)
}

func (cs clusterStore) ZRem(key string, values ...interface{}) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRem(key, values...)
}

func (cs clusterStore) ZRemRangeByRank(key string, start, stop int64) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRemRangeByRank(key, start, stop)
}

func (cs clusterStore) ZRemRangeByScore(key string, start, stop int64) (int, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRemRangeByScore(key, start, stop)
}

func (cs clusterStore) ZRevRange(key string, start, stop int64) ([]string, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRevRange(key, start, stop)
}

func (cs clusterStore) ZRevRangeByScoreWithScores(key string, start, stop int64) ([]redis.Pair, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRevRangeByScoreWithScores(key, start, stop)
}

func (cs clusterStore) ZRevRangeByScoreWithScoresAndLimit(key string, start, stop int64, page, size int) (
	[]redis.Pair, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return nil, err
	}

	return node.ZRevRangeByScoreWithScoresAndLimit(key, start, stop, page, size)
}

func (cs clusterStore) ZRevRank(key, field string) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZRevRank(key, field)
}

func (cs clusterStore) ZScore(key string, value string) (int64, error) {
	node, err := cs.getRedis(key)
	if err != nil {
		return 0, err
	}

	return node.ZScore(key, value)
}

func (cs clusterStore) getRedis(key string) (*redis.Redis, error) {
	if val, ok := cs.dispatcher.Get(key); !ok {
		return nil, ErrNoRedisNode
	} else {
		return val.(*redis.Redis), nil
	}
}
