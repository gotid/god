package cache

import (
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/store/redis/redistest"
	"github.com/gotid/god/lib/syncx"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

var errTestNotFound = errors.New("测试环境未找到")

func init() {
	logx.Disable()
	stat.SetReporter(nil)
}

func TestNode_Del(t *testing.T) {
	rds, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	rds.Type = redis.ClusterType
	defer clean()

	n := node{
		rds:            rds,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errTestNotFound,
	}

	assert.Nil(t, n.Del())
	assert.Nil(t, n.Del([]string{}...))
	assert.Nil(t, n.Del(make([]string, 0)...))
	n.Set("first", "one")
	assert.Nil(t, n.Del("first"))
	n.Set("first", "one")
	n.Set("second", "two")
	assert.Nil(t, n.Del("first", "second"))
}

func TestNode_DelWithErrors(t *testing.T) {
	rds, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	rds.Type = redis.ClusterType
	defer clean()

	n := node{
		rds:            rds,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errTestNotFound,
	}

	assert.Nil(t, n.Del("first", "second"))
}

func TestNode_InvalidCache(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	n := node{
		rds:            redis.New(mr.Addr()),
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errTestNotFound,
	}

	err = mr.Set("any", "value")
	var str string
	err = n.Get("any", &str)
	assert.NotNil(t, err)
	assert.Equal(t, "", str)
	_, err = mr.Get("any")
	assert.Equal(t, miniredis.ErrKeyNotFound, err)
}

func TestCacheNode_InvalidCache(t *testing.T) {
	mr, err := miniredis.Run()
	assert.Nil(t, err)
	defer mr.Close()

	cn := node{
		rds:            redis.New(mr.Addr()),
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errTestNotFound,
	}

	err = mr.Set("any", "value")
	var str string
	err = cn.Get("any", &str)
	assert.NotNil(t, err)
	assert.Equal(t, "", str)
	_, err = mr.Get("any")
	assert.Equal(t, miniredis.ErrKeyNotFound, err)
}

func TestCacheNode_SetWithExpire(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	cn := node{
		rds:            store,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		barrier:        syncx.NewSingleFlight(),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errors.New("any"),
	}
	assert.NotNil(t, cn.SetWithExpire("key", make(chan int), time.Second))
}

func TestCacheNode_Take(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	n := NewNode(store, syncx.NewSingleFlight(), NewStat("any"), errTestNotFound,
		WithExpire(time.Second), WithNotFoundExpire(time.Second))
	var str string
	err = n.Take(&str, "any", func(v any) error {
		*v.(*string) = "value"
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "value", str)
	assert.Nil(t, n.Get("any", &str))
	val, err := store.Get("any")
	assert.Nil(t, err)
	assert.Equal(t, `"value"`, val)
}

func TestCacheNode_TakeNotFound(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	cn := node{
		rds:            store,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		barrier:        syncx.NewSingleFlight(),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errTestNotFound,
	}
	var str string
	err = cn.Take(&str, "any", func(v any) error {
		return errTestNotFound
	})
	assert.True(t, cn.IsNotFound(err))
	assert.True(t, cn.IsNotFound(cn.Get("any", &str)))
	val, err := store.Get("any")
	assert.Nil(t, err)
	assert.Equal(t, `*`, val)

	store.Set("any", "*")
	err = cn.Take(&str, "any", func(v any) error {
		return nil
	})
	assert.True(t, cn.IsNotFound(err))
	assert.True(t, cn.IsNotFound(cn.Get("any", &str)))

	store.Del("any")
	errDummy := errors.New("dummy")
	err = cn.Take(&str, "any", func(v any) error {
		return errDummy
	})
	assert.Equal(t, errDummy, err)
}

func TestCacheNode_TakeWithExpire(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	n := node{
		rds:            store,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		barrier:        syncx.NewSingleFlight(),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errors.New("any"),
	}
	var str string
	err = n.TakeWithExpire(&str, "any", func(v any, expire time.Duration) error {
		*v.(*string) = "value"
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "value", str)
	assert.Nil(t, n.Get("any", &str))
	val, err := store.Get("any")
	assert.Nil(t, err)
	assert.Equal(t, `"value"`, val)
}

func TestCacheNode_String(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	n := node{
		rds:            store,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		barrier:        syncx.NewSingleFlight(),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errors.New("any"),
	}
	assert.Equal(t, store.Addr, n.String())
}

func TestCacheValueWithBigInt(t *testing.T) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	cn := node{
		rds:            store,
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		barrier:        syncx.NewSingleFlight(),
		lock:           new(sync.Mutex),
		unstableExpire: mathx.NewUnstable(expireDeviation),
		stat:           NewStat("any"),
		errNotFound:    errors.New("any"),
	}

	const (
		key         = "key"
		value int64 = 323427211229009810
	)

	assert.Nil(t, cn.Set(key, value))
	var val any
	assert.Nil(t, cn.Get(key, &val))
	assert.Equal(t, strconv.FormatInt(value, 10), fmt.Sprintf("%v", val))
}
