package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/errorx"
	"github.com/gotid/god/lib/hash"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/store/redis/redistest"
	"github.com/gotid/god/lib/syncx"
	"github.com/stretchr/testify/assert"
	"math"
	"strconv"
	"testing"
	"time"
)

var _ Cache = (*mockedNode)(nil)

type mockedNode struct {
	values      map[string][]byte
	errNotFound error
}

func (mn *mockedNode) Del(keys ...string) error {
	return mn.DelCtx(context.Background(), keys...)
}

func (mn *mockedNode) DelCtx(ctx context.Context, keys ...string) error {
	var be errorx.BatchError

	for _, key := range keys {
		if _, ok := mn.values[key]; !ok {
			be.Add(mn.errNotFound)
		} else {
			delete(mn.values, key)
		}
	}

	return be.Err()
}

func (mn *mockedNode) Get(key string, val interface{}) error {
	return mn.GetCtx(context.Background(), key, val)
}

func (mn *mockedNode) GetCtx(ctx context.Context, key string, val interface{}) error {
	bs, ok := mn.values[key]
	if ok {
		return json.Unmarshal(bs, val)
	}

	return mn.errNotFound
}

func (mn *mockedNode) IsNotFound(err error) bool {
	return errors.Is(err, mn.errNotFound)
}

func (mn *mockedNode) Set(key string, val interface{}) error {
	return mn.SetCtx(context.Background(), key, val)
}

func (mn *mockedNode) SetCtx(ctx context.Context, key string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	mn.values[key] = data
	return nil
}

func (mn *mockedNode) SetWithExpire(key string, val interface{}, expire time.Duration) error {
	return mn.SetWithExpireCtx(context.Background(), key, val, expire)
}

func (mn *mockedNode) SetWithExpireCtx(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	return mn.Set(key, val)
}

func (mn *mockedNode) Take(val interface{}, key string, query func(val interface{}) error) error {
	return mn.TakeCtx(context.Background(), val, key, query)
}

func (mn *mockedNode) TakeCtx(ctx context.Context, val interface{}, key string, query func(val interface{}) error) error {
	if _, ok := mn.values[key]; ok {
		return mn.GetCtx(ctx, key, val)
	}

	if err := query(val); err != nil {
		return err
	}

	return mn.SetCtx(ctx, key, val)
}

func (mn *mockedNode) TakeWithExpire(val interface{}, key string, query func(val interface{}, expire time.Duration) error) error {
	return mn.TakeWithExpireCtx(context.Background(), val, key, query)
}

func (mn *mockedNode) TakeWithExpireCtx(ctx context.Context, val interface{}, key string, query func(val interface{}, expire time.Duration) error) error {
	return mn.Take(val, key, func(val interface{}) error {
		return query(val, 0)
	})
}

func TestCache_SetDel(t *testing.T) {
	const total = 1000
	r1, clean1, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean1()
	r2, clean2, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean2()

	conf := ClusterConfig{
		{
			Config: redis.Config{
				Host: r1.Addr,
				Type: redis.NodeType,
			},
			Weight: 100,
		},
		{
			Config: redis.Config{
				Host: r2.Addr,
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	}
	c := New(conf, syncx.NewSingleFlight(), NewStat("mock"), errPlaceholder)
	for i := 0; i < total; i++ {
		if i%2 == 0 {
			assert.Nil(t, c.Set(fmt.Sprintf("key/%d", i), i))
		} else {
			assert.Nil(t, c.SetWithExpire(fmt.Sprintf("key/%d", i), i, 0))
		}
	}
	for i := 0; i < total; i++ {
		var val int
		assert.Nil(t, c.Get(fmt.Sprintf("key/%d", i), &val))
		assert.Equal(t, i, val)
	}
	assert.Nil(t, c.Del())
	for i := 0; i < total; i++ {
		assert.Nil(t, c.Del(fmt.Sprintf("key/%d", i)))
	}
	for i := 0; i < total; i++ {
		var val int
		assert.True(t, c.IsNotFound(c.Get(fmt.Sprintf("key/%d", i), &val)))
		assert.Equal(t, 0, val)
	}
}

func TestCache_OneNode(t *testing.T) {
	const total = 1000
	r, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()
	conf := ClusterConfig{
		{
			Config: redis.Config{
				Host: r.Addr,
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	}
	c := New(conf, syncx.NewSingleFlight(), NewStat("mock"), errPlaceholder)
	for i := 0; i < total; i++ {
		if i%2 == 0 {
			assert.Nil(t, c.Set(fmt.Sprintf("key/%d", i), i))
		} else {
			assert.Nil(t, c.SetWithExpire(fmt.Sprintf("key/%d", i), i, 0))
		}
	}
	for i := 0; i < total; i++ {
		var val int
		assert.Nil(t, c.Get(fmt.Sprintf("key/%d", i), &val))
		assert.Equal(t, i, val)
	}
	assert.Nil(t, c.Del())
	for i := 0; i < total; i++ {
		assert.Nil(t, c.Del(fmt.Sprintf("key/%d", i)))
	}
	for i := 0; i < total; i++ {
		var val int
		assert.True(t, c.IsNotFound(c.Get(fmt.Sprintf("key/%d", i), &val)))
		assert.Equal(t, 0, val)
	}
}

func TestCache_Balance(t *testing.T) {
	const (
		numNodes = 100
		total    = 10000
	)
	dispatcher := hash.NewConsistentHash()
	maps := make([]map[string][]byte, numNodes)
	for i := 0; i < numNodes; i++ {
		maps[i] = map[string][]byte{
			strconv.Itoa(i): []byte(strconv.Itoa(i)),
		}
	}
	for i := 0; i < numNodes; i++ {
		dispatcher.AddWithWeight(&mockedNode{
			values:      maps[i],
			errNotFound: errPlaceholder,
		}, 100)
	}

	c := cluster{
		dispatcher:  dispatcher,
		errNotFound: errPlaceholder,
	}
	for i := 0; i < total; i++ {
		assert.Nil(t, c.Set(strconv.Itoa(i), i))
	}

	counts := make(map[int]int)
	for i, m := range maps {
		counts[i] = len(m)
	}
	entropy := calcEntropy(counts, total)
	assert.True(t, len(counts) > 1)
	assert.True(t, entropy > .95, fmt.Sprintf("entropy should be greater than 0.95, but got %.2f", entropy))

	for i := 0; i < total; i++ {
		var val int
		assert.Nil(t, c.Get(strconv.Itoa(i), &val))
		assert.Equal(t, i, val)
	}

	for i := 0; i < total/10; i++ {
		assert.Nil(t, c.Del(strconv.Itoa(i*10), strconv.Itoa(i*10+1), strconv.Itoa(i*10+2)))
		assert.Nil(t, c.Del(strconv.Itoa(i*10+9)))
	}

	var count int
	for i := 0; i < total/10; i++ {
		var val int
		if i%2 == 0 {
			assert.Nil(t, c.Take(&val, strconv.Itoa(i*10), func(val interface{}) error {
				*val.(*int) = i
				count++
				return nil
			}))
		} else {
			assert.Nil(t, c.TakeWithExpire(&val, strconv.Itoa(i*10), func(val interface{}, expire time.Duration) error {
				*val.(*int) = i
				count++
				return nil
			}))
		}
		assert.Equal(t, i, val)
	}
	assert.Equal(t, total/10, count)
}

func TestCacheNoNode(t *testing.T) {
	dispatcher := hash.NewConsistentHash()
	c := cluster{
		dispatcher:  dispatcher,
		errNotFound: errPlaceholder,
	}
	assert.NotNil(t, c.Del("foo"))
	assert.NotNil(t, c.Del("foo", "bar", "any"))
	assert.NotNil(t, c.Get("foo", nil))
	assert.NotNil(t, c.Set("foo", nil))
	assert.NotNil(t, c.SetWithExpire("foo", nil, time.Second))
	assert.NotNil(t, c.Take(nil, "foo", func(val interface{}) error {
		return nil
	}))
	assert.NotNil(t, c.TakeWithExpire(nil, "foo", func(val interface{}, duration time.Duration) error {
		return nil
	}))
}

func calcEntropy(m map[int]int, total int) float64 {
	var entropy float64

	for _, val := range m {
		proba := float64(val) / float64(total)
		entropy -= proba * math.Log2(proba)
	}

	return entropy / math.Log2(float64(len(m)))
}
