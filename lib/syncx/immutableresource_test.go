package syncx

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestImmutableResource(t *testing.T) {
	var count int
	ir := NewImmutableResource(func() (any, error) {
		fmt.Println("请求资源")
		count++
		return "hellox", nil
	})

	res, err := ir.Get()
	assert.Nil(t, err)
	assert.Equal(t, "hellox", res)
	assert.Equal(t, 1, count)

	// 再来一次
	res, err = ir.Get()
	assert.Nil(t, err)
	assert.Equal(t, "hellox", res)
	assert.Equal(t, 1, count)
}

func TestImmutableResource_Error(t *testing.T) {
	var count int
	ir := NewImmutableResource(func() (any, error) {
		count++
		return nil, errors.New("any")
	})

	res, err := ir.Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "any", err.Error())
	assert.Equal(t, 1, count)

	// 再来一次
	res, err = ir.Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "any", err.Error())
	assert.Equal(t, 1, count)

	ir.refreshInterval = 0
	time.Sleep(time.Millisecond)
	res, err = ir.Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "any", err.Error())
	assert.Equal(t, 2, count)
}

func TestImmutableResource_ErrorRefreshAlways(t *testing.T) {
	var count int
	ir := NewImmutableResource(func() (any, error) {
		count++
		return nil, errors.New("any")
	}, WithRefreshIntervalOnFailure(0))

	res, err := ir.Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "any", err.Error())
	assert.Equal(t, 1, count)

	// 再来一次
	res, err = ir.Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "any", err.Error())
	assert.Equal(t, 2, count)
}
