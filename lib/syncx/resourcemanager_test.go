package syncx

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type dummyResource struct {
	age int
}

func (r *dummyResource) Close() error {
	return errors.New("close")
}

func TestResourceManager_Get(t *testing.T) {
	m := NewResourceManager()
	defer m.Close()

	var age int
	for i := 0; i < 10; i++ {
		val, err := m.Get("key", func() (io.Closer, error) {
			age++
			return &dummyResource{
				age: age,
			}, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 1, val.(*dummyResource).age)
	}
}

func TestResourceManager_GetResourceError(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	for i := 0; i < 10; i++ {
		_, err := manager.Get("key", func() (io.Closer, error) {
			return nil, errors.New("fail")
		})
		assert.NotNil(t, err)
	}
}

func TestResourceManager_Close(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	for i := 0; i < 10; i++ {
		_, err := manager.Get("key", func() (io.Closer, error) {
			return nil, errors.New("fail")
		})
		assert.NotNil(t, err)
	}

	if assert.NoError(t, manager.Close()) {
		assert.Equal(t, 0, len(manager.resources))
	}
}

func TestResourceManager_UseAfterClose(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	_, err := manager.Get("key", func() (io.Closer, error) {
		return nil, errors.New("fail")
	})
	assert.NotNil(t, err)
	if assert.NoError(t, manager.Close()) {
		_, err = manager.Get("key", func() (io.Closer, error) {
			return nil, errors.New("fail")
		})
		assert.NotNil(t, err)

		assert.Panics(t, func() {
			_, err = manager.Get("key", func() (io.Closer, error) {
				return &dummyResource{age: 123}, nil
			})
		})
	}
}

func TestResourceManager_Set(t *testing.T) {
	manager := NewResourceManager()
	defer manager.Close()

	manager.Set("key", &dummyResource{
		age: 10,
	})

	val, err := manager.Get("key", func() (io.Closer, error) {
		return nil, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 10, val.(*dummyResource).age)
}
