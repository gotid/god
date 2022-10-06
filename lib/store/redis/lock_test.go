package redis

import (
	"context"
	"github.com/gotid/god/lib/stringx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLock(t *testing.T) {
	testFn := func(ctx context.Context) func(client *Redis) {
		return func(client *Redis) {
			key := stringx.Rand()
			firstLock := NewLock(client, key)
			firstLock.SetExpire(5)
			firstAcquire, err := firstLock.Acquire()
			assert.Nil(t, err)
			assert.True(t, firstAcquire)

			secondLock := NewLock(client, key)
			secondLock.SetExpire(5)
			againAcquire, err := secondLock.Acquire()
			assert.Nil(t, err)
			assert.False(t, againAcquire)

			release, err := firstLock.Release()
			assert.Nil(t, err)
			assert.True(t, release)

			endAcquire, err := secondLock.Acquire()
			assert.Nil(t, err)
			assert.True(t, endAcquire)
		}
	}

	t.Run("normal", func(t *testing.T) {
		runOnRedis(t, testFn(nil))
	})

	t.Run("withContext", func(t *testing.T) {
		runOnRedis(t, testFn(context.Background()))
	})
}

func TestLock_Expire(t *testing.T) {
	runOnRedis(t, func(client *Redis) {
		key := stringx.Rand()
		lock := NewLock(client, key)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := lock.AcquireCtx(ctx)
		assert.NotNil(t, err)
	})

	runOnRedis(t, func(client *Redis) {
		key := stringx.Rand()
		lock := NewLock(client, key)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := lock.ReleaseCtx(ctx)
		assert.NotNil(t, err)
	})
}
