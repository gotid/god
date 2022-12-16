package redis

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/alicebob/miniredis/v2"
	red "github.com/go-redis/redis/v8"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stringx"
	"github.com/stretchr/testify/assert"
	"io"
	"strconv"
	"testing"
	"time"
)

func TestRedis_Decr(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).Decr("a")
		assert.NotNil(t, err)
		val, err := client.Decr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), val)
		val, err = client.Decr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(-2), val)
	})
}

func TestRedis_DecrBy(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).DecrBy("a", 2)
		assert.NotNil(t, err)
		val, err := client.DecrBy("a", 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(-2), val)
		val, err = client.DecrBy("a", 3)
		assert.Nil(t, err)
		assert.Equal(t, int64(-5), val)
	})
}

func TestRedis_Exists(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).Exists("a")
		assert.NotNil(t, err)
		ok, err := client.Exists("a")
		assert.Nil(t, err)
		assert.False(t, ok)
		assert.Nil(t, client.Set("a", "b"))
		ok, err = client.Exists("a")
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedisTLS_Exists(t *testing.T) {
	runOnRedisTLS(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).Exists("a")
		assert.NotNil(t, err)
		ok, err := client.Exists("a")
		assert.NotNil(t, err)
		assert.False(t, ok)
		assert.NotNil(t, client.Set("a", "b"))
		ok, err = client.Exists("a")
		assert.NotNil(t, err)
		assert.False(t, ok)
	})
}

func TestRedis_Eval(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).Eval(`redis.call("EXISTS", KEYS[1])`, []string{"notexist"})
		assert.NotNil(t, err)
		_, err = client.Eval(`redis.call("EXISTS", KEYS[1])`, []string{"notexist"})
		assert.Equal(t, Nil, err)
		err = client.Set("key1", "value1")
		assert.Nil(t, err)
		_, err = client.Eval(`redis.call("EXISTS", KEYS[1])`, []string{"key1"})
		assert.Equal(t, Nil, err)
		val, err := client.Eval(`return redis.call("EXISTS", KEYS[1])`, []string{"key1"})
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
	})
}

func TestRedis_GeoHash(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := client.GeoHash("parent", "child1", "child2")
		assert.NotNil(t, err)
	})
}

func TestRedis_Hgetall(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		_, err := New(client.Addr, badType()).HGetAll("a")
		assert.NotNil(t, err)
		vals, err := client.HGetAll("a")
		assert.Nil(t, err)
		assert.EqualValues(t, map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}, vals)
	})
}

func TestRedis_Hvals(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.NotNil(t, New(client.Addr, badType()).HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		_, err := New(client.Addr, badType()).HVals("a")
		assert.NotNil(t, err)
		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestRedis_Hsetnx(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		_, err := New(client.Addr, badType()).HSetNX("a", "bb", "ccc")
		assert.NotNil(t, err)
		ok, err := client.HSetNX("a", "bb", "ccc")
		assert.Nil(t, err)
		assert.False(t, ok)
		ok, err = client.HSetNX("a", "dd", "ddd")
		assert.Nil(t, err)
		assert.True(t, ok)
		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb", "ddd"}, vals)
	})
}

func TestRedis_HdelHlen(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		_, err := New(client.Addr, badType()).HLen("a")
		assert.NotNil(t, err)
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
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).HIncrBy("key", "field", 2)
		assert.NotNil(t, err)
		val, err := client.HIncrBy("key", "field", 2)
		assert.Nil(t, err)
		assert.Equal(t, 2, val)
		val, err = client.HIncrBy("key", "field", 3)
		assert.Nil(t, err)
		assert.Equal(t, 5, val)
	})
}

func TestRedis_Hkeys(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		_, err := New(client.Addr, badType()).HKeys("a")
		assert.NotNil(t, err)
		vals, err := client.HKeys("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aa", "bb"}, vals)
	})
}

func TestRedis_Hmget(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		_, err := New(client.Addr, badType()).HMGet("a", "aa", "bb")
		assert.NotNil(t, err)
		vals, err := client.HMGet("a", "aa", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "bbb"}, vals)
		vals, err = client.HMGet("a", "aa", "no", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "", "bbb"}, vals)
	})
}

func TestRedis_Hmset(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.NotNil(t, New(client.Addr, badType()).HMSet("a", nil))
		assert.Nil(t, client.HMSet("a", map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}))
		vals, err := client.HMGet("a", "aa", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestRedis_Hscan(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		key := "hash:hi"
		fieldsAndValues := make(map[string]string)
		for i := 0; i < 1550; i++ {
			fieldsAndValues["filed_"+strconv.Itoa(i)] = stringx.Randn(i)
		}
		err := client.HMSet(key, fieldsAndValues)
		assert.Nil(t, err)

		var cursor uint64 = 0
		sum := 0
		for {
			_, _, err := New(client.Addr, badType()).HScan(key, cursor, "*", 100)
			assert.NotNil(t, err)
			reMap, next, err := client.HScan(key, cursor, "*", 100)
			assert.Nil(t, err)
			sum += len(reMap)
			if next == 0 {
				break
			}
			cursor = next
		}

		assert.Equal(t, sum, 3100)
		_, err = New(client.Addr, badType()).Del(key)
		assert.NotNil(t, err)
		_, err = client.Del(key)
		assert.Nil(t, err)
	})
}

func TestRedis_Incr(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).Incr("a")
		assert.NotNil(t, err)
		val, err := client.Incr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		val, err = client.Incr("a")
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
	})
}

func TestRedis_IncrBy(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).IncrBy("a", 2)
		assert.NotNil(t, err)
		val, err := client.IncrBy("a", 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		val, err = client.IncrBy("a", 3)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
	})
}

func TestRedis_Keys(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "value1")
		assert.Nil(t, err)
		err = client.Set("key2", "value2")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).Keys("*")
		assert.NotNil(t, err)
		keys, err := client.Keys("*")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
	})
}

func TestRedis_HyperLogLog(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		r := New(client.Addr)
		ok, err := r.PFAdd("key1", "val1")
		assert.Nil(t, err)
		assert.True(t, ok)
		val, err := r.PFCount("key1")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		ok, err = r.PFAdd("key2", "val2")
		assert.Nil(t, err)
		assert.True(t, ok)
		val, err = r.PFCount("key2")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		err = r.PFMerge("key1", "key2")
		assert.Nil(t, err)
		val, err = r.PFCount("key1")
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
	})
}

func TestRedis_List(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).LPush("key", "value1", "value2")
		assert.NotNil(t, err)
		val, err := client.LPush("key", "value1", "value2")
		assert.Nil(t, err)
		assert.Equal(t, 2, val)
		_, err = New(client.Addr, badType()).RPush("key", "value3", "value4")
		assert.NotNil(t, err)
		val, err = client.RPush("key", "value3", "value4")
		assert.Nil(t, err)
		assert.Equal(t, 4, val)
		_, err = New(client.Addr, badType()).LLen("key")
		assert.NotNil(t, err)
		val, err = client.LLen("key")
		assert.Nil(t, err)
		assert.Equal(t, 4, val)
		_, err = New(client.Addr, badType()).LIndex("key", 1)
		assert.NotNil(t, err)
		value, err := client.LIndex("key", 0)
		assert.Nil(t, err)
		assert.Equal(t, "value2", value)
		vals, err := client.LRange("key", 0, 10)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value1", "value3", "value4"}, vals)
		_, err = New(client.Addr, badType()).LPop("key")
		assert.NotNil(t, err)
		v, err := client.LPop("key")
		assert.Nil(t, err)
		assert.Equal(t, "value2", v)
		val, err = client.LPush("key", "value1", "value2")
		assert.Nil(t, err)
		assert.Equal(t, 5, val)
		_, err = New(client.Addr, badType()).RPop("key")
		assert.NotNil(t, err)
		v, err = client.RPop("key")
		assert.Nil(t, err)
		assert.Equal(t, "value4", v)
		val, err = client.RPush("key", "value4", "value3", "value3")
		assert.Nil(t, err)
		assert.Equal(t, 7, val)
		_, err = New(client.Addr, badType()).LRem("key", 2, "value1")
		assert.NotNil(t, err)
		n, err := client.LRem("key", 2, "value1")
		assert.Nil(t, err)
		assert.Equal(t, 2, n)
		_, err = New(client.Addr, badType()).LRange("key", 0, 10)
		assert.NotNil(t, err)
		vals, err = client.LRange("key", 0, 10)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value3", "value4", "value3", "value3"}, vals)
		n, err = client.LRem("key", -2, "value3")
		assert.Nil(t, err)
		assert.Equal(t, 2, n)
		vals, err = client.LRange("key", 0, 10)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value3", "value4"}, vals)
		err = client.LTrim("key", 0, 1)
		assert.Nil(t, err)
		vals, err = client.LRange("key", 0, 10)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value2", "value3"}, vals)
	})
}

func TestRedis_Mget(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "value1")
		assert.Nil(t, err)
		err = client.Set("key2", "value2")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).MGet("key1", "key0", "key2", "key3")
		assert.NotNil(t, err)
		vals, err := client.MGet("key1", "key0", "key2", "key3")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value1", "", "value2", ""}, vals)
	})
}

func TestRedis_SetBit(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).SetBit("key", 1, 1)
		assert.NotNil(t, err)
		val, err := client.SetBit("key", 1, 1)
		assert.Nil(t, err)
		assert.Equal(t, 0, val)
	})
}

func TestRedis_GetBit(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		val, err := client.SetBit("key", 2, 1)
		assert.Nil(t, err)
		assert.Equal(t, 0, val)
		_, err = New(client.Addr, badType()).GetBit("key", 2)
		assert.NotNil(t, err)
		v, err := client.GetBit("key", 2)
		assert.Nil(t, err)
		assert.Equal(t, 1, v)
	})
}

func TestRedis_BitCount(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		for i := 0; i < 11; i++ {
			val, err := client.SetBit("key", int64(i), 1)
			assert.Nil(t, err)
			assert.Equal(t, 0, val)
		}

		_, err := New(client.Addr, badType()).BitCount("key", 0, -1)
		assert.NotNil(t, err)
		val, err := client.BitCount("key", 0, -1)
		assert.Nil(t, err)
		assert.Equal(t, int64(11), val)

		val, err = client.BitCount("key", 0, 0)
		assert.Nil(t, err)
		assert.Equal(t, int64(8), val)

		val, err = client.BitCount("key", 1, 1)
		assert.Nil(t, err)
		assert.Equal(t, int64(3), val)

		val, err = client.BitCount("key", 0, 1)
		assert.Nil(t, err)
		assert.Equal(t, int64(11), val)

		val, err = client.BitCount("key", 2, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), val)
	})
}

func TestRedis_BitOpAnd(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "0")
		assert.Nil(t, err)
		err = client.Set("key2", "1")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).BitOpAnd("destKey", "key1", "key2")
		assert.NotNil(t, err)
		val, err := client.BitOpAnd("destKey", "key1", "key2")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		valStr, err := client.Get("destKey")
		assert.Nil(t, err)
		// destKey  binary 110000   ascii 0
		assert.Equal(t, "0", valStr)
	})
}

func TestRedis_BitOpNot(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "\u0000")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).BitOpNot("destKey", "key1")
		assert.NotNil(t, err)
		val, err := client.BitOpNot("destKey", "key1")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		valStr, err := client.Get("destKey")
		assert.Nil(t, err)
		assert.Equal(t, "\xff", valStr)
	})
}

func TestRedis_BitOpOr(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "1")
		assert.Nil(t, err)
		err = client.Set("key2", "0")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).BitOpOr("destKey", "key1", "key2")
		assert.NotNil(t, err)
		val, err := client.BitOpOr("destKey", "key1", "key2")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		valStr, err := client.Get("destKey")
		assert.Nil(t, err)
		assert.Equal(t, "1", valStr)
	})
}

func TestRedis_BitOpXor(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "\xff")
		assert.Nil(t, err)
		err = client.Set("key2", "\x0f")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).BitOpXor("destKey", "key1", "key2")
		assert.NotNil(t, err)
		val, err := client.BitOpXor("destKey", "key1", "key2")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), val)
		valStr, err := client.Get("destKey")
		assert.Nil(t, err)
		assert.Equal(t, "\xf0", valStr)
	})
}

func TestRedis_BitPos(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		// 11111111 11110000 00000000
		err := client.Set("key", "\xff\xf0\x00")
		assert.Nil(t, err)

		_, err = New(client.Addr, badType()).BitPos("key", 0, 0, -1)
		assert.NotNil(t, err)
		val, err := client.BitPos("key", 0, 0, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(12), val)

		val, err = client.BitPos("key", 1, 0, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(0), val)

		val, err = client.BitPos("key", 0, 1, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(12), val)

		val, err = client.BitPos("key", 1, 1, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(8), val)

		val, err = client.BitPos("key", 1, 2, 2)
		assert.Nil(t, err)
		assert.Equal(t, int64(-1), val)
	})
}

func TestRedis_Persist(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).Persist("key")
		assert.NotNil(t, err)
		ok, err := client.Persist("key")
		assert.Nil(t, err)
		assert.False(t, ok)
		err = client.Set("key", "value")
		assert.Nil(t, err)
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.False(t, ok)
		err = New(client.Addr, badType()).Expire("key", 5)
		assert.NotNil(t, err)
		err = client.Expire("key", 5)
		assert.Nil(t, err)
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.True(t, ok)
		err = New(client.Addr, badType()).ExpireAt("key", time.Now().Unix()+5)
		assert.NotNil(t, err)
		err = client.ExpireAt("key", time.Now().Unix()+5)
		assert.Nil(t, err)
		ok, err = client.Persist("key")
		assert.Nil(t, err)
		assert.True(t, ok)
	})
}

func TestRedis_Ping(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		ok := client.Ping()
		assert.True(t, ok)
	})
}

func TestRedis_Scan(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.Set("key1", "value1")
		assert.Nil(t, err)
		err = client.Set("key2", "value2")
		assert.Nil(t, err)
		_, _, err = New(client.Addr, badType()).Scan(0, "*", 100)
		assert.NotNil(t, err)
		keys, _, err := client.Scan(0, "*", 100)
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
	})
}

func TestRedis_Sscan(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		key := "list"
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
			_, _, err := New(client.Addr, badType()).SScan(key, cursor, "", 100)
			assert.NotNil(t, err)
			keys, next, err := client.SScan(key, cursor, "", 100)
			assert.Nil(t, err)
			sum += len(keys)
			if next == 0 {
				break
			}
			cursor = next
		}

		assert.Equal(t, sum, 1550)
		_, err = New(client.Addr, badType()).Del(key)
		assert.NotNil(t, err)
		_, err = client.Del(key)
		assert.Nil(t, err)
	})
}

func TestRedis_Set(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).SAdd("key", 1, 2, 3, 4)
		assert.NotNil(t, err)
		num, err := client.SAdd("key", 1, 2, 3, 4)
		assert.Nil(t, err)
		assert.Equal(t, 4, num)
		_, err = New(client.Addr, badType()).SCard("key")
		assert.NotNil(t, err)
		val, err := client.SCard("key")
		assert.Nil(t, err)
		assert.Equal(t, int64(4), val)
		_, err = New(client.Addr, badType()).SIsMember("key", 2)
		assert.NotNil(t, err)
		ok, err := client.SIsMember("key", 2)
		assert.Nil(t, err)
		assert.True(t, ok)
		_, err = New(client.Addr, badType()).SRem("key", 3, 4)
		assert.NotNil(t, err)
		num, err = client.SRem("key", 3, 4)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		_, err = New(client.Addr, badType()).SMembers("key")
		assert.NotNil(t, err)
		vals, err := client.SMembers("key")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"1", "2"}, vals)
		_, err = New(client.Addr, badType()).SRandMember("key", 1)
		assert.NotNil(t, err)
		members, err := client.SRandMember("key", 1)
		assert.Nil(t, err)
		assert.Len(t, members, 1)
		assert.Contains(t, []string{"1", "2"}, members[0])
		_, err = New(client.Addr, badType()).SPop("key")
		assert.NotNil(t, err)
		member, err := client.SPop("key")
		assert.Nil(t, err)
		assert.Contains(t, []string{"1", "2"}, member)
		_, err = New(client.Addr, badType()).SMembers("key")
		assert.NotNil(t, err)
		vals, err = client.SMembers("key")
		assert.Nil(t, err)
		assert.NotContains(t, vals, member)
		_, err = New(client.Addr, badType()).SAdd("key1", 1, 2, 3, 4)
		assert.NotNil(t, err)
		num, err = client.SAdd("key1", 1, 2, 3, 4)
		assert.Nil(t, err)
		assert.Equal(t, 4, num)
		num, err = client.SAdd("key2", 2, 3, 4, 5)
		assert.Nil(t, err)
		assert.Equal(t, 4, num)
		_, err = New(client.Addr, badType()).SUnion("key1", "key2")
		assert.NotNil(t, err)
		vals, err = client.SUnion("key1", "key2")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"1", "2", "3", "4", "5"}, vals)
		_, err = New(client.Addr, badType()).SUnionStore("key3", "key1", "key2")
		assert.NotNil(t, err)
		num, err = client.SUnionStore("key3", "key1", "key2")
		assert.Nil(t, err)
		assert.Equal(t, 5, num)
		_, err = New(client.Addr, badType()).SDiff("key1", "key2")
		assert.NotNil(t, err)
		vals, err = client.SDiff("key1", "key2")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"1"}, vals)
		_, err = New(client.Addr, badType()).SDiffStore("key4", "key1", "key2")
		assert.NotNil(t, err)
		num, err = client.SDiffStore("key4", "key1", "key2")
		assert.Nil(t, err)
		assert.Equal(t, 1, num)
		_, err = New(client.Addr, badType()).SInter("key1", "key2")
		assert.NotNil(t, err)
		vals, err = client.SInter("key1", "key2")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"2", "3", "4"}, vals)
		_, err = New(client.Addr, badType()).SInterStore("key4", "key1", "key2")
		assert.NotNil(t, err)
		num, err = client.SInterStore("key4", "key1", "key2")
		assert.Nil(t, err)
		assert.Equal(t, 3, num)
	})
}

func TestRedis_GetSet(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		_, err := New(client.Addr, badType()).GetSet("hello", "world")
		assert.NotNil(t, err)
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
		ret, err := client.Del("hello")
		assert.Nil(t, err)
		assert.Equal(t, 1, ret)
	})
}

func TestRedis_SetGetDel(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := New(client.Addr, badType()).Set("hello", "world")
		assert.NotNil(t, err)
		err = client.Set("hello", "world")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).Get("hello")
		assert.NotNil(t, err)
		val, err := client.Get("hello")
		assert.Nil(t, err)
		assert.Equal(t, "world", val)
		ret, err := client.Del("hello")
		assert.Nil(t, err)
		assert.Equal(t, 1, ret)
	})
}

func TestRedis_SetExNx(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := New(client.Addr, badType()).SetEx("hello", "world", 5)
		assert.NotNil(t, err)
		err = client.SetEx("hello", "world", 5)
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).SetNX("hello", "newworld")
		assert.NotNil(t, err)
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
		_, err = New(client.Addr, badType()).SetNXEx("newhello", "newworld", 5)
		assert.NotNil(t, err)
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

func TestRedis_SetGetDelHashField(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := client.HSet("key", "field", "value")
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).HGet("key", "field")
		assert.NotNil(t, err)
		val, err := client.HGet("key", "field")
		assert.Nil(t, err)
		assert.Equal(t, "value", val)
		_, err = New(client.Addr, badType()).HExists("key", "field")
		assert.NotNil(t, err)
		ok, err := client.HExists("key", "field")
		assert.Nil(t, err)
		assert.True(t, ok)
		_, err = New(client.Addr, badType()).HDel("key", "field")
		assert.NotNil(t, err)
		ret, err := client.HDel("key", "field")
		assert.Nil(t, err)
		assert.True(t, ret)
		ok, err = client.HExists("key", "field")
		assert.Nil(t, err)
		assert.False(t, ok)
	})
}

func TestRedis_SortedSet(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		ok, err := client.ZAddFloat("key", 1, "value1")
		assert.Nil(t, err)
		assert.True(t, ok)
		ok, err = client.ZAdd("key", 2, "value1")
		assert.Nil(t, err)
		assert.False(t, ok)
		val, err := client.ZScore("key", "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		_, err = New(client.Addr, badType()).ZIncrBy("key", 3, "value1")
		assert.NotNil(t, err)
		val, err = client.ZIncrBy("key", 3, "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
		_, err = New(client.Addr, badType()).ZScore("key", "value1")
		assert.NotNil(t, err)
		val, err = client.ZScore("key", "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(5), val)
		_, err = New(client.Addr, badType()).ZAdds("key")
		assert.NotNil(t, err)
		val, err = client.ZAdds("key", Pair{
			Member: "value2",
			Score:  6,
		}, Pair{
			Member: "value3",
			Score:  7,
		})
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		_, err = New(client.Addr, badType()).ZRevRangeWithScores("key", 1, 3)
		assert.NotNil(t, err)
		pairs, err := client.ZRevRangeWithScores("key", 1, 3)
		assert.Nil(t, err)
		assert.EqualValues(t, []Pair{
			{
				Member: "value2",
				Score:  6,
			},
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		rank, err := client.ZRank("key", "value2")
		assert.Nil(t, err)
		assert.Equal(t, int64(1), rank)
		rank, err = client.ZRevRank("key", "value1")
		assert.Nil(t, err)
		assert.Equal(t, int64(2), rank)
		_, err = New(client.Addr, badType()).ZRank("key", "value4")
		assert.NotNil(t, err)
		_, err = client.ZRank("key", "value4")
		assert.Equal(t, Nil, err)
		_, err = New(client.Addr, badType()).ZRem("key", "value2", "value3")
		assert.NotNil(t, err)
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
		_, err = New(client.Addr, badType()).ZRemRangeByScore("key", 6, 7)
		assert.NotNil(t, err)
		num, err = client.ZRemRangeByScore("key", 6, 7)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		ok, err = client.ZAdd("key", 6, "value2")
		assert.Nil(t, err)
		assert.True(t, ok)
		_, err = New(client.Addr, badType()).ZAdd("key", 7, "value3")
		assert.NotNil(t, err)
		ok, err = client.ZAdd("key", 7, "value3")
		assert.Nil(t, err)
		assert.True(t, ok)
		_, err = New(client.Addr, badType()).ZCount("key", 6, 7)
		assert.NotNil(t, err)
		num, err = client.ZCount("key", 6, 7)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		_, err = New(client.Addr, badType()).ZRemRangeByRank("key", 1, 2)
		assert.NotNil(t, err)
		num, err = client.ZRemRangeByRank("key", 1, 2)
		assert.Nil(t, err)
		assert.Equal(t, 2, num)
		_, err = New(client.Addr, badType()).ZCard("key")
		assert.NotNil(t, err)
		card, err := client.ZCard("key")
		assert.Nil(t, err)
		assert.Equal(t, 2, card)
		_, err = New(client.Addr, badType()).ZRange("key", 0, -1)
		assert.NotNil(t, err)
		vals, err := client.ZRange("key", 0, -1)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value1", "value4"}, vals)
		_, err = New(client.Addr, badType()).ZRevRange("key", 0, -1)
		assert.NotNil(t, err)
		vals, err = client.ZRevRange("key", 0, -1)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"value4", "value1"}, vals)
		_, err = New(client.Addr, badType()).ZRangeWithScores("key", 0, -1)
		assert.NotNil(t, err)
		pairs, err = client.ZRangeWithScores("key", 0, -1)
		assert.Nil(t, err)
		assert.EqualValues(t, []Pair{
			{
				Member: "value1",
				Score:  5,
			},
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		_, err = New(client.Addr, badType()).ZRangeByScoreWithScores("key", 5, 8)
		assert.NotNil(t, err)
		pairs, err = client.ZRangeByScoreWithScores("key", 5, 8)
		assert.Nil(t, err)
		assert.EqualValues(t, []Pair{
			{
				Member: "value1",
				Score:  5,
			},
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		_, err = New(client.Addr, badType()).ZRangeByScoreWithScoresAndLimit(
			"key", 5, 8, 1, 1)
		assert.NotNil(t, err)
		pairs, err = client.ZRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 1)
		assert.Nil(t, err)
		assert.EqualValues(t, []Pair{
			{
				Member: "value4",
				Score:  8,
			},
		}, pairs)
		pairs, err = client.ZRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 0)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(pairs))
		_, err = New(client.Addr, badType()).ZRevRangeByScoreWithScores("key", 5, 8)
		assert.NotNil(t, err)
		pairs, err = client.ZRevRangeByScoreWithScores("key", 5, 8)
		assert.Nil(t, err)
		assert.EqualValues(t, []Pair{
			{
				Member: "value4",
				Score:  8,
			},
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		_, err = New(client.Addr, badType()).ZRevRangeByScoreWithScoresAndLimit(
			"key", 5, 8, 1, 1)
		assert.NotNil(t, err)
		pairs, err = client.ZRevRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 1)
		assert.Nil(t, err)
		assert.EqualValues(t, []Pair{
			{
				Member: "value1",
				Score:  5,
			},
		}, pairs)
		pairs, err = client.ZRevRangeByScoreWithScoresAndLimit("key", 5, 8, 1, 0)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(pairs))
		_, err = New(client.Addr, badType()).ZRevRank("key", "value")
		assert.NotNil(t, err)
		_, _ = client.ZAdd("second", 2, "aa")
		_, _ = client.ZAdd("third", 3, "bbb")
		val, err = client.ZUnionStore("union", &ZStore{
			Keys:      []string{"second", "third"},
			Weights:   []float64{1, 2},
			Aggregate: "SUM",
		})
		assert.Nil(t, err)
		assert.Equal(t, int64(2), val)
		_, err = New(client.Addr, badType()).ZUnionStore("union", &ZStore{})
		assert.NotNil(t, err)
		vals, err = client.ZRange("union", 0, 10000)
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aa", "bbb"}, vals)
		ival, err := client.ZCard("union")
		assert.Nil(t, err)
		assert.Equal(t, 2, ival)
	})
}

func TestRedis_Pipelined(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		assert.NotNil(t, New(client.Addr, badType()).Pipelined(func(pipeliner Pipeliner) error {
			return nil
		}))
		err := client.Pipelined(
			func(pipe Pipeliner) error {
				pipe.Incr(context.Background(), "pipelined_counter")
				pipe.Expire(context.Background(), "pipelined_counter", time.Hour)
				pipe.ZAdd(context.Background(), "zadd", &Z{Score: 12, Member: "zadd"})
				return nil
			},
		)
		assert.Nil(t, err)
		_, err = New(client.Addr, badType()).TTL("pipelined_counter")
		assert.NotNil(t, err)
		ttl, err := client.TTL("pipelined_counter")
		assert.Nil(t, err)
		assert.Equal(t, 3600, ttl)
		value, err := client.Get("pipelined_counter")
		assert.Nil(t, err)
		assert.Equal(t, "1", value)
		score, err := client.ZScore("zadd", "zadd")
		assert.Nil(t, err)
		assert.Equal(t, int64(12), score)
	})
}

func TestRedisString(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		rc := New(client.Addr, WithCluster())
		_, err := getRedis(rc)
		assert.Nil(t, err)
		assert.Equal(t, client.Addr, client.String())
		assert.NotNil(t, New(client.Addr, badType()).Ping())
	})
}

func TestRedisScriptLoad(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		_, err := New(client.Addr, badType()).ScriptLoad("foo")
		assert.NotNil(t, err)
		_, err = client.ScriptLoad("foo")
		assert.NotNil(t, err)
	})
}

func TestRedisEvalSha(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		scriptHash, err := client.ScriptLoad(`return redis.call("EXISTS", KEYS[1])`)
		assert.Nil(t, err)
		result, err := client.EvalSha(scriptHash, []string{"key1"})
		assert.Nil(t, err)
		assert.Equal(t, int64(0), result)
	})
}

func TestRedisToPairs(t *testing.T) {
	pairs := toPairs([]red.Z{
		{
			Member: 1,
			Score:  1,
		},
		{
			Member: 2,
			Score:  2,
		},
	})
	assert.EqualValues(t, []Pair{
		{
			Member: "1",
			Score:  1,
		},
		{
			Member: "2",
			Score:  2,
		},
	}, pairs)
}

func TestRedisToStrings(t *testing.T) {
	vals := toStrings([]any{1, 2})
	assert.EqualValues(t, []string{"1", "2"}, vals)
}

func TestRedisBlpop(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		var node mockedNode
		_, err := client.BLPop(nil, "foo")
		assert.NotNil(t, err)
		_, err = client.BLPop(node, "foo")
		assert.NotNil(t, err)
	})
}

func TestRedisBlpopEx(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		var node mockedNode
		_, _, err := client.BLPopEx(nil, "foo")
		assert.NotNil(t, err)
		_, _, err = client.BLPopEx(node, "foo")
		assert.NotNil(t, err)
	})
}

func TestRedisBlpopWithTimeout(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		var node mockedNode
		_, err := client.BLPopWithTimeout(nil, 10*time.Second, "foo")
		assert.NotNil(t, err)
		_, err = client.BLPopWithTimeout(node, 10*time.Second, "foo")
		assert.NotNil(t, err)
	})
}

func TestRedisGeo(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		client.Ping()
		geoLocation := []*GeoLocation{{Longitude: 13.361389, Latitude: 38.115556, Name: "Palermo"}, {Longitude: 15.087269, Latitude: 37.502669, Name: "Catania"}}
		v, err := client.GeoAdd("sicily", geoLocation...)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), v)
		v2, err := client.GeoDist("sicily", "Palermo", "Catania", "m")
		assert.Nil(t, err)
		assert.Equal(t, 166274, int(v2))
		// GeoHash not support
		v3, err := client.GeoPos("sicily", "Palermo", "Catania")
		assert.Nil(t, err)
		assert.Equal(t, int64(v3[0].Longitude), int64(13))
		assert.Equal(t, int64(v3[0].Latitude), int64(38))
		assert.Equal(t, int64(v3[1].Longitude), int64(15))
		assert.Equal(t, int64(v3[1].Latitude), int64(37))
		v4, err := client.GeoRadius("sicily", 15, 37, &red.GeoRadiusQuery{WithDist: true, Unit: "km", Radius: 200})
		assert.Nil(t, err)
		assert.Equal(t, int64(v4[0].Dist), int64(190))
		assert.Equal(t, int64(v4[1].Dist), int64(56))
		geoLocation2 := []*GeoLocation{{Longitude: 13.583333, Latitude: 37.316667, Name: "Agrigento"}}
		v5, err := client.GeoAdd("sicily", geoLocation2...)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), v5)
		v6, err := client.GeoRadiusByMember("sicily", "Agrigento", &red.GeoRadiusQuery{Unit: "km", Radius: 100})
		assert.Nil(t, err)
		assert.Equal(t, v6[0].Name, "Agrigento")
		assert.Equal(t, v6[1].Name, "Palermo")
	})
}

func TestSetSlowThreshold(t *testing.T) {
	assert.Equal(t, defaultSlowThreshold, slowThreshold.Load())
	SetSlowThreshold(time.Second)
	assert.Equal(t, time.Second, slowThreshold.Load())
}

func TestRedis_WithPass(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		err := New(client.Addr, WithPass("any")).Ping()
		assert.NotNil(t, err)
	})
}

func runOnRedis(t *testing.T, fn func(client *Redis)) {
	logx.Disable()

	s, err := miniredis.Run()
	assert.Nil(t, err)
	defer func() {
		client, err := clientManager.Get(s.Addr(), func() (io.Closer, error) {
			return nil, errors.New("已经存在了，不应该再创建")
		})
		if err != nil {
			t.Error(err)
		}

		if client != nil {
			_ = client.Close()
		}
	}()
	fn(New(s.Addr()))
}

func runOnRedisTLS(t *testing.T, fn func(client *Redis)) {
	logx.Disable()

	s, err := miniredis.RunTLS(&tls.Config{
		Certificates:       make([]tls.Certificate, 1),
		InsecureSkipVerify: true,
	})
	assert.Nil(t, err)
	defer func() {
		client, err := clientManager.Get(s.Addr(), func() (io.Closer, error) {
			return nil, errors.New("已经存在了，不应该再创建")
		})
		if err != nil {
			t.Error(err)
		}
		if client != nil {
			_ = client.Close()
		}
	}()
	fn(New(s.Addr(), WithTLS()))
}

func badType() Option {
	return func(r *Redis) {
		r.Type = "bad"
	}
}

type mockedNode struct {
	Node
}

func (n mockedNode) BLPop(_ context.Context, _ time.Duration, _ ...string) *red.StringSliceCmd {
	return red.NewStringSliceCmd(context.Background(), "foo", "bar")
}
