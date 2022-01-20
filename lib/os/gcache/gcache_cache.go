package gcache

import (
	"time"

	"github.com/gotid/god/lib/container/gvar"
	"github.com/gotid/god/lib/gconv"
	"github.com/gotid/god/lib/os/gtimer"
)

// Cache struct.
type Cache struct {
	Adapter // Adapter for cache features.
}

// New creates and returns a new cache object using default memory adapter.
// Note that the LRU feature is only available using memory adapter.
func New(lruCap ...int) *Cache {
	memAdapter := newAdapterMemory(lruCap...)
	c := &Cache{
		Adapter: memAdapter,
	}
	// Here may be a "timer leak" if adapter is manually changed from memory adapter.
	// Do not worry about this, as adapter is less changed and it dose nothing if it's not used.
	gtimer.AddSingleton(time.Second, memAdapter.syncEventAndClearExpired)
	return c
}

// SetAdapter changes the adapter for this cache.
// Be very note that, this setting function is not concurrent-safe, which means you should not call
// this setting function concurrently in multiple goroutines.
func (c *Cache) SetAdapter(adapter Adapter) {
	c.Adapter = adapter
}

// GetVar retrieves and returns the value of <key> as gvar.Var.
func (c *Cache) GetVar(key interface{}) (*gvar.Var, error) {
	v, err := c.Get(key)
	return gvar.New(v), err
}

// Removes deletes <keys> in the cache.
// Deprecated, use Remove instead.
func (c *Cache) Removes(keys []interface{}) error {
	_, err := c.Remove(keys...)
	return err
}

// KeyStrings returns all keys in the cache as string slice.
func (c *Cache) KeyStrings() ([]string, error) {
	keys, err := c.Keys()
	if err != nil {
		return nil, err
	}
	return gconv.Strings(keys), nil
}
