package redis

import (
	"sync"
	"sync/atomic"
)

var (
	once        sync.Once
	lock        sync.Mutex
	scriptCache *ScriptCache
)

type (
	// Map 是 map[string]string 的简称。
	Map map[string]string

	// ScriptCache 是一种缓存，用来存储带有 sha1 键的脚本。
	ScriptCache struct {
		atomic.Value
	}
)

// GetScriptCache 获取一个脚本缓存实例 ScriptCache。
func GetScriptCache() *ScriptCache {
	once.Do(func() {
		scriptCache = &ScriptCache{}
		scriptCache.Store(make(Map))
	})

	return scriptCache
}

// GetSha 获取给定脚本的 sha 校验码。
func (c *ScriptCache) GetSha(script string) (string, bool) {
	cache := c.Load().(Map)
	ret, ok := cache[script]
	return ret, ok
}

// SetSha 设置给定脚本的校验码为 sha。
func (c *ScriptCache) SetSha(script, sha string) {
	lock.Lock()
	defer lock.Unlock()

	cache := c.Load().(Map)
	newCache := make(Map)
	for k, v := range cache {
		newCache[k] = v
	}
	newCache[script] = sha
	c.Store(newCache)
}
