package cache

import "time"

// Cache 接口定义了缓存所需实现的方法。
type Cache interface {
	Del(keys ...string) error
	Get(key string, dest interface{}) error
	MGet(keys []string, dest []interface{}) error
	Set(key string, val interface{}) error
	SetEx(key string, val interface{}, expires time.Duration) error
	SetBit(key string, offset int64, value int) error
	SetBits(key string, offset []int64) error
	UnsetBits(key string, offset []int64) error
	GetBit(key string, offset int64) (int, error)
	GetBits(key string, offset []int64) (map[int64]bool, error)
	Take(dest interface{}, key string, queryFn func(interface{}) error) error
	TakeEx(dest interface{}, key string, queryFn func(interface{}, time.Duration) error) error
}
