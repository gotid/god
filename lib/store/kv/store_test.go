package kv

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/gotid/god/lib/hash"
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/stringx"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
)

func TestRedis_Decr(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.Decr("a")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		val, err := client.Decr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), val)
		val, err = client.Decr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(-2), val)
	})
}

func TestRedis_DecrBy(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.IncrBy("a", 2)
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		val, err := client.DecrBy("a", 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(-2), val)
		val, err = client.DecrBy("a", 3)
		assert.Nil(t, err)
		assert.Equal(t, int64(-5), val)
	})
}

func TestRedis_Exists(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.Exists("foo")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		ok, err := client.Exists("a")
		assert.Nil(t, err)
		assert.False(t, ok)
		assert.Nil(t, client.Set("a", "b"))
		ok, err = client.Exists("a")
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedis_Eval(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.Eval(`redis.call("EXISTS", KEYS[1])`, "key1")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		_, err := client.Eval(`redis.call("EXISTS", KEYS[1])`, "notexist")
		assert.Equal(t, redis.Nil, err)
		err = client.Set("key1", "value1")
		assert.Nil(t, err)
		_, err = client.Eval(`redis.call("EXISTS", KEYS[1])`, "key1")
		assert.Equal(t, redis.Nil, err)
		val, err := client.Eval(`return redis.call("EXISTS", KEYS[1])`, "key1")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
	})
}

func TestRedis_Hgetall(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	err := store.HSet("a", "aa", "aaa")
	assert.NotNil(t, err)
	_, err = store.HGetAll("a")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		vals, err := client.HGetAll("a")
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}, vals)
	})
}

func TestRedis_Hvals(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HVals("a")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestRedis_Hsetnx(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HSetNx("a", "dd", "ddd")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		ok, err := client.HSetNx("a", "bb", "ccc")
		assert.Nil(t, err)
		assert.False(t, ok)
		ok, err = client.HSetNx("a", "dd", "ddd")
		assert.Nil(t, err)
		assert.True(t, ok)
		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb", "ddd"}, vals)
	})
}

func TestRedis_HdelHlen(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HDel("a", "aa")
	assert.NotNil(t, err)
	_, err = store.HLen("a")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		num, err := client.HLen("a")
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		val, err := client.HDel("a", "aa")
		assert.Nil(t, err)
		assert.True(t, val)
		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"bbb"}, vals)
	})
}

func TestRedis_HIncrBy(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HIncrBy("key", "field", 3)
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		val, err := client.HIncrBy("key", "field", 2)
		assert.Nil(t, err)
		assert.Equal(t, 2, val)
		val, err = client.HIncrBy("key", "field", 3)
		assert.Nil(t, err)
		assert.Equal(t, 5, val)
	})
}

func TestRedis_Hkeys(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HKeys("a")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		vals, err := client.HKeys("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aa", "bb"}, vals)
	})
}

func TestRedis_Hmget(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HMGet("a", "aa", "bb")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		vals, err := client.HMGet("a", "aa", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "bbb"}, vals)
		vals, err = client.HMGet("a", "aa", "no", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "", "bbb"}, vals)
	})
}

func TestRedis_Hmset(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	err := store.HMSet("a", map[string]string{
		"aa": "aaa",
	})
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		assert.Nil(t, client.HMSet("a", map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}))
		vals, err := client.HMGet("a", "aa", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestRedis_Incr(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.Incr("a")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		val, err := client.Incr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		val, err = client.Incr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
	})
}

func TestRedis_IncrBy(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.IncrBy("a", 2)
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		val, err := client.IncrBy("a", 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		val, err = client.IncrBy("a", 3)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
	})
}

func TestRedis_List(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.LPush("key", "value1", "value2")
	assert.NotNil(t, err)
	_, err = store.RPush("key", "value3", "value4")
	assert.NotNil(t, err)
	_, err = store.LLen("key")
	assert.NotNil(t, err)
	_, err = store.LRange("key", 0, 10)
	assert.NotNil(t, err)
	_, err = store.LPop("key")
	assert.NotNil(t, err)
	_, err = store.LRem("key", 0, "val")
	assert.NotNil(t, err)
	_, err = store.LIndex("key", 0)
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		val, err := client.LPush("key", "value1", "value2")
		assert.Nil(t, err)
		assert.Equal(t, 2, val)
		val, err = client.RPush("key", "value3", "value4")
		assert.Nil(t, err)
		assert.Equal(t, 4, val)
		val, err = client.LLen("key")
		assert.Nil(t, err)
		assert.Equal(t, 4, val)
		value, err := client.LIndex("key", 0)
		assert.Nil(t, err)
		assert.Equal(t, "value2", value)
		vals, err := client.LRange("key", 0, 10)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value1", "value3", "value4"}, vals)
		v, err := client.LPop("key")
		assert.Nil(t, err)
		assert.Equal(t, "value2", v)
		val, err = client.LPush("key", "value1", "value2")
		assert.Nil(t, err)
		assert.Equal(t, 5, val)
		val, err = client.RPush("key", "value3", "value3")
		assert.Nil(t, err)
		assert.Equal(t, 7, val)
		n, err := client.LRem("key", 2, "value1")
		assert.Nil(t, err)
		assert.Equal(t, 2, n)
		vals, err = client.LRange("key", 0, 10)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value3", "value4", "value3", "value3"}, vals)
		n, err = client.LRem("key", -2, "value3")
		assert.Nil(t, err)
		assert.Equal(t, 2, n)
		vals, err = client.LRange("key", 0, 10)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value3", "value4"}, vals)
	})
}

func TestRedis_Persist(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.Persist("key")
	assert.NotNil(t, err)
	err = store.Expire("key", 5)
	assert.NotNil(t, err)
	err = store.ExpireAt("key", time.Now().Unix()+5)
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		ok, err := client.Persist("key")
		assert.Nil(t, err)
		assert.False(t, ok)
		err = client.Set("key", "value")
		assert.Nil(t, err)
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.False(t, ok)
		err = client.Expire("key", 5)
		assert.Nil(t, err)
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.True(t, ok)
		err = client.ExpireAt("key", time.Now().Unix()+5)
		assert.Nil(t, err)
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedis_Sscan(t *testing.T) {
	key := "list"
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.SAdd(key, nil)
	assert.NotNil(t, err)
	_, _, err = store.SScan(key, 0, "", 100)
	assert.NotNil(t, err)
	_, err = store.Del(key)
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		var list []string
		for i := 0; i < 1550; i++ {
			list = append(list, stringx.Randn(i))
		}
		lens, err := client.SAdd(key, list)
		assert.Nil(t, err)
		assert.Equal(t, lens, 1550)

		var cursor uint64 = 0
		sum := 0
		for {
			keys, next, err := client.SScan(key, cursor, "", 100)
			assert.Nil(t, err)
			sum += len(keys)
			if next == 0 {
				break
			}
			cursor = next
		}

		assert.Equal(t, sum, 1550)
		_, err = client.Del(key)
		assert.Nil(t, err)
	})
}

func TestRedis_Set(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.SCard("key")
	assert.NotNil(t, err)
	_, err = store.SIsMember("key", 2)
	assert.NotNil(t, err)
	_, err = store.SRem("key", 3, 4)
	assert.NotNil(t, err)
	_, err = store.SMembers("key")
	assert.NotNil(t, err)
	_, err = store.SRandMember("key", 1)
	assert.NotNil(t, err)
	_, err = store.SPop("key")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		num, err := client.SAdd("key", 1, 2, 3, 4)
		assert.Nil(t, err)
		assert.Equal(t, 4, num)
		val, err := client.SCard("key")
		assert.Nil(t, err)
		assert.Equal(t, int64(4), val)
		ok, err := client.SIsMember("key", 2)
		assert.Nil(t, err)
		assert.True(t, ok)
		num, err = client.SRem("key", 3, 4)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		vals, err := client.SMembers("key")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"1", "2"}, vals)
		members, err := client.SRandMember("key", 1)
		assert.Nil(t, err)
		assert.Len(t, members, 1)
		assert.Contains(t, []string{"1", "2"}, members[0])
		member, err := client.SPop("key")
		assert.Nil(t, err)
		assert.Contains(t, []string{"1", "2"}, member)
		vals, err = client.SMembers("key")
		assert.Nil(t, err)
		assert.NotContains(t, vals, member)
		num, err = client.SAdd("key1", 1, 2, 3, 4)
		assert.Nil(t, err)
		assert.Equal(t, 4, num)
		num, err = client.SAdd("key2", 2, 3, 4, 5)
		assert.Nil(t, err)
		assert.Equal(t, 4, num)
	})
}

func TestRedis_SetGetDel(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	err := store.Set("hello", "world")
	assert.NotNil(t, err)
	_, err = store.Get("hello")
	assert.NotNil(t, err)
	_, err = store.Del("hello")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		err := client.Set("hello", "world")
		assert.Nil(t, err)
		val, err := client.Get("hello")
		assert.Nil(t, err)
		assert.Equal(t, "world", val)
		ret, err := client.Del("hello")
		assert.Nil(t, err)
		assert.Equal(t, 1, ret)
	})
}

func TestRedis_SetExNx(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	err := store.SetEx("hello", "world", 5)
	assert.NotNil(t, err)
	_, err = store.SetNX("newhello", "newworld")
	assert.NotNil(t, err)
	_, err = store.TTL("hello")
	assert.NotNil(t, err)
	_, err = store.SetNXEx("newhello", "newworld", 5)
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		err := client.SetEx("hello", "world", 5)
		assert.Nil(t, err)
		ok, err := client.SetNX("hello", "newworld")
		assert.Nil(t, err)
		assert.False(t, ok)
		ok, err = client.SetNX("newhello", "newworld")
		assert.Nil(t, err)
		assert.True(t, ok)
		val, err := client.Get("hello")
		assert.Nil(t, err)
		assert.Equal(t, "world", val)
		val, err = client.Get("newhello")
		assert.Nil(t, err)
		assert.Equal(t, "newworld", val)
		ttl, err := client.TTL("hello")
		assert.Nil(t, err)
		assert.True(t, ttl > 0)
		ok, err = client.SetNXEx("newhello", "newworld", 5)
		assert.Nil(t, err)
		assert.False(t, ok)
		num, err := client.Del("newhello")
		assert.Nil(t, err)
		assert.Equal(t, 1, num)
		ok, err = client.SetNXEx("newhello", "newworld", 5)
		assert.Nil(t, err)
		assert.True(t, ok)
		val, err = client.Get("newhello")
		assert.Nil(t, err)
		assert.Equal(t, "newworld", val)
	})
}

func TestRedis_Getset(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.GetSet("hello", "world")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		val, err := client.GetSet("hello", "world")
		assert.Nil(t, err)
		assert.Equal(t, "", val)
		val, err = client.Get("hello")
		assert.Nil(t, err)
		assert.Equal(t, "world", val)
		val, err = client.GetSet("hello", "newworld")
		assert.Nil(t, err)
		assert.Equal(t, "world", val)
		val, err = client.Get("hello")
		assert.Nil(t, err)
		assert.Equal(t, "newworld", val)
		_, err = client.Del("hello")
		assert.Nil(t, err)
	})
}

func TestRedis_SetGetDelHashField(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	err := store.HSet("key", "field", "value")
	assert.NotNil(t, err)
	_, err = store.HGet("key", "field")
	assert.NotNil(t, err)
	_, err = store.HExists("key", "field")
	assert.NotNil(t, err)
	_, err = store.HDel("key", "field")
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		err := client.HSet("key", "field", "value")
		assert.Nil(t, err)
		val, err := client.HGet("key", "field")
		assert.Nil(t, err)
		assert.Equal(t, "value", val)
		ok, err := client.HExists("key", "field")
		assert.Nil(t, err)
		assert.True(t, ok)
		ret, err := client.HDel("key", "field")
		assert.Nil(t, err)
		assert.True(t, ret)
		ok, err = client.HExists("key", "field")
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestRedis_SortedSet(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.ZAdd("key", 1, "value1")
	assert.NotNil(t, err)
	_, err = store.ZScore("key", "value1")
	assert.NotNil(t, err)
	_, err = store.ZCount("key", 6, 7)
	assert.NotNil(t, err)
	_, err = store.ZIncrBy("key", 3, "value1")
	assert.NotNil(t, err)
	_, err = store.ZRank("key", "value2")
	assert.NotNil(t, err)
	_, err = store.ZRem("key", "value2", "value3")
	assert.NotNil(t, err)
	_, err = store.ZRemRangeByScore("key", 6, 7)
	assert.NotNil(t, err)
	_, err = store.ZRemRangeByRank("key", 1, 2)
	assert.NotNil(t, err)
	_, err = store.ZCard("key")
	assert.NotNil(t, err)
	_, err = store.ZRange("key", 0, -1)
	assert.NotNil(t, err)
	_, err = store.ZRevRange("key", 0, -1)
	assert.NotNil(t, err)
	_, err = store.ZRangeWithScores("key", 0, -1)
	assert.NotNil(t, err)
	_, err = store.ZRangeByScoreWithScores("key", 5, 8)
	assert.NotNil(t, err)
	_, err = store.ZRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 1)
	assert.NotNil(t, err)
	_, err = store.ZRevRangeByScoreWithScores("key", 5, 8)
	assert.NotNil(t, err)
	_, err = store.ZRevRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 1)
	assert.NotNil(t, err)
	_, err = store.ZRevRank("key", "value")
	assert.NotNil(t, err)
	_, err = store.ZAdds("key", redis.Pair{
		Member: "value2",
		Score:  6,
	}, redis.Pair{
		Member: "value3",
		Score:  7,
	})
	assert.NotNil(t, err)

	runOnCluster(func(client Store) {
		ok, err := client.ZAddFloat("key", 1, "value1")
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = client.ZAdd("key", 2, "value1")
		assert.Nil(t, err)
		assert.False(t, ok)
		val, err := client.ZScore("key", "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		val, err = client.ZIncrBy("key", 3, "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
		val, err = client.ZScore("key", "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
		ok, err = client.ZAdd("key", 6, "value2")
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = client.ZAdd("key", 7, "value3")
		assert.Nil(t, err)
		assert.True(t, ok)
		rank, err := client.ZRank("key", "value2")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), rank)
		_, err = client.ZRank("key", "value4")
		assert.Equal(t, redis.Nil, err)
		num, err := client.ZRem("key", "value2", "value3")
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		ok, err = client.ZAdd("key", 6, "value2")
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = client.ZAdd("key", 7, "value3")
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = client.ZAdd("key", 8, "value4")
		assert.Nil(t, err)
		assert.True(t, ok)
		num, err = client.ZRemRangeByScore("key", 6, 7)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		ok, err = client.ZAdd("key", 6, "value2")
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = client.ZAdd("key", 7, "value3")
		assert.Nil(t, err)
		assert.True(t, ok)
		num, err = client.ZCount("key", 6, 7)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		num, err = client.ZRemRangeByRank("key", 1, 2)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		card, err := client.ZCard("key")
		assert.Nil(t, err)
		assert.Equal(t, 2, card)
		vals, err := client.ZRange("key", 0, -1)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value1", "value4"}, vals)
		vals, err = client.ZRevRange("key", 0, -1)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value4", "value1"}, vals)
		pairs, err := client.ZRangeWithScores("key", 0, -1)
		assert.Nil(t, err)
		assert.EqualValues(t, []redis.Pair{
			{
				Member: "value1",
				Score:  5,
			},
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		pairs, err = client.ZRangeByScoreWithScores("key", 5, 8)
		assert.Nil(t, err)
		assert.EqualValues(t, []redis.Pair{
			{
				Member: "value1",
				Score:  5,
			},
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		pairs, err = client.ZRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 1)
		assert.Nil(t, err)
		assert.EqualValues(t, []redis.Pair{
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		pairs, err = client.ZRevRangeByScoreWithScores("key", 5, 8)
		assert.Nil(t, err)
		assert.EqualValues(t, []redis.Pair{
			{
				Member: "value4",
				Score:  8,
			},
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		pairs, err = client.ZRevRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 1)
		assert.Nil(t, err)
		assert.EqualValues(t, []redis.Pair{
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		rank, err = client.ZRevRank("key", "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), rank)
		val, err = client.ZAdds("key", redis.Pair{
			Member: "value2",
			Score:  6,
		}, redis.Pair{
			Member: "value3",
			Score:  7,
		})
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
	})
}

func TestRedis_HyperLogLog(t *testing.T) {
	store := kvStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.PFAdd("key")
	assert.NotNil(t, err)
	_, err = store.PFCount("key")
	assert.NotNil(t, err)

	runOnCluster(func(cluster Store) {
		ok, err := cluster.PFAdd("key", "a", "b", "a", "c")
		assert.Nil(t, err)
		assert.True(t, ok)
		val, err := cluster.PFCount("key")
		assert.Nil(t, err)
		assert.Equal(t, int64(3), val)
	})
}

func runOnCluster(fn func(cluster Store)) {
	s1.FlushAll()
	s2.FlushAll()

	store := New([]cache.NodeConfig{
		{
			Config: redis.Config{
				Host: s1.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
		{
			Config: redis.Config{
				Host: s2.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	})

	fn(store)
}
