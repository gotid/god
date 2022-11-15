package limit

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/store/redis/redistest"
	"github.com/stretchr/testify/assert"
)

func TestPeriodLimit_Take(t *testing.T) {
	testPeriodLimit(t)
}

func TestPeriodLimit_TakeWithAlign(t *testing.T) {
	testPeriodLimit(t, Align())
}

func TestPeriodLimit_RedisUnavailable(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	const (
		seconds = 1
		quota   = 5
	)
	l := NewPeriodLimit(seconds, quota, redis.New(s.Addr()), "sendSmsLimit")
	s.Close()
	val, err := l.Take("uid10010")
	assert.NotNil(t, err)
	assert.Equal(t, 0, val)
}

func testPeriodLimit(t *testing.T, opts ...PeriodOption) {
	store, clean, err := redistest.CreateRedis()
	assert.Nil(t, err)
	defer clean()

	const (
		seconds = 1
		quota   = 5
		total   = 100
	)

	l := NewPeriodLimit(seconds, quota, store, "sendSmsLimit", opts...)
	var allowed, hitQuota, overQuota int
	for i := 0; i < total; i++ {
		val, err := l.Take("uid10010")
		if err != nil {
			t.Error(err)
		}
		switch val {
		case Allowed:
			allowed++
		case HitQuota:
			hitQuota++
		case OverQuota:
			overQuota++
		default:
			t.Error("未知状态")
		}
	}

	assert.Equal(t, quota-1, allowed)
	assert.Equal(t, 1, hitQuota)
	assert.Equal(t, total-quota, overQuota)
}

func TestQuotaFull(t *testing.T) {
	s, err := miniredis.Run()
	assert.Nil(t, err)

	// 1秒只能1发次，上来就达到限额
	l := NewPeriodLimit(1, 1, redis.New(s.Addr()), "sendSmsLimit")
	val, err := l.Take("uid10010")
	assert.Nil(t, err)
	assert.Equal(t, HitQuota, val)
}
