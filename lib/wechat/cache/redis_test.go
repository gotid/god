package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.zc0901.com/go/god/lib/store/cache"
	"git.zc0901.com/go/god/lib/store/kv"
	"git.zc0901.com/go/god/lib/store/redis"
)

func TestRedis(t *testing.T) {
	store := kv.NewStore([]cache.Conf{
		{
			Conf: redis.Conf{
				Host:     "vps:6382",
				Password: "4a5d4787a82c660ee18719f51ff40d9a669a4958",
				Mode:     redis.StandaloneMode,
			},
			Weight: 100,
		},
		{
			Conf: redis.Conf{
				Host:     "vps:6382",
				Password: "4a5d4787a82c660ee18719f51ff40d9a669a4958",
				Mode:     redis.StandaloneMode,
			},
			Weight: 100,
		},
	})
	r := NewRedis(store)

	err := r.Set("username", "god", 1*time.Second)
	assert.Nil(t, err)
	assert.True(t, r.Exists("username"))
	assert.Equal(t, "god", r.Get("username"))
	err = r.Delete("username")
	assert.Nil(t, err)
}
