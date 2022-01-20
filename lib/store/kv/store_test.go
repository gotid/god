package kv

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gotid/god/lib/hash"
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/redis"
	"github.com/stretchr/testify/assert"
)

var (
	s1, _ = miniredis.Run()
	s2, _ = miniredis.Run()
)

func runOnCluster(t *testing.T, fn func(cluster Store)) {
	s1.FlushAll()
	s2.FlushAll()

	store := NewStore([]cache.Conf{
		{
			Conf: redis.Conf{
				Host: s1.Addr(),
				Mode: redis.StandaloneMode,
			},
			Weight: 100,
		},
		{
			Conf: redis.Conf{
				Host: s2.Addr(),
				Mode: redis.StandaloneMode,
			},
			Weight: 100,
		},
	})

	fn(store)
}

func TestRedis_MGet(t *testing.T) {
	runOnCluster(t, func(client Store) {
		assert.Nil(t, client.Set("a", "1"))
		assert.Nil(t, client.Set("b", "2"))

		vals, err := client.MGet("a", "b", "c")
		assert.Nil(t, err)
		//assert.EqualValues(t, map[string]string{
		//	"aa": "aaa",
		//	"bb": "bbb",
		//}, vals)
		fmt.Println(vals)
		fmt.Println(client.Get("a"))
		fmt.Println(client.Get("b"))
	})
}

func TestRedis_Exists(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.Exists("foo")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
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
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.Eval(`redis.call("EXISTS", KEYS[1])`, "key1")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
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

func TestRedis_HGetAll(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	err := store.HSet("a", "aa", "aaa")
	assert.NotNil(t, err)
	_, err = store.HGetAll("a")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
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

func TestRedis_HVals(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HVals("a")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		vals, err := client.HVals("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestRedis_HSetNX(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HSetNX("a", "dd", "ddd")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
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

func TestRedis_HDelHLen(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HDel("a", "aa")
	assert.NotNil(t, err)
	_, err = store.HLen("a")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
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
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HIncrBy("key", "field", 3)
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
		val, err := client.HIncrBy("key", "field", 2)
		assert.Nil(t, err)
		assert.Equal(t, 2, val)
		val, err = client.HIncrBy("key", "field", 3)
		assert.Nil(t, err)
		assert.Equal(t, 5, val)
	})
}

func TestRedis_HKeys(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HKeys("a")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
		assert.Nil(t, client.HSet("a", "aa", "aaa"))
		assert.Nil(t, client.HSet("a", "bb", "bbb"))
		vals, err := client.HKeys("a")
		assert.Nil(t, err)
		assert.ElementsMatch(t, []string{"aa", "bb"}, vals)
	})
}

func TestRedis_HMGet(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	_, err := store.HMGet("a", "aa", "bb")
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
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

func TestRedis_HMSet(t *testing.T) {
	store := clusterStore{dispatcher: hash.NewConsistentHash()}
	err := store.HMSet("a", map[string]string{
		"aa": "aaa",
	})
	assert.NotNil(t, err)

	runOnCluster(t, func(client Store) {
		assert.Nil(t, client.HMSet("a", map[string]string{
			"aa": "aaa",
			"bb": "bbb",
		}))
		vals, err := client.HMGet("a", "aa", "bb")
		assert.Nil(t, err)
		assert.EqualValues(t, []string{"aaa", "bbb"}, vals)
	})
}

func TestClusterStore_SetBit(t *testing.T) {
	runOnCluster(t, func(store Store) {
		for i := 0; i < 10; i++ {
			value := rand.Intn(2)
			fmt.Printf("INPUT %d\n", value)
			err := store.SetBit("test", int64(i), value)
			assert.Nil(t, err)
		}

		for i := 0; i < 10; i++ {
			result, err := store.GetBit("test", int64(i))
			assert.Nil(t, err)
			fmt.Printf("OUTPUT %d\n", result)
		}
	})
}
