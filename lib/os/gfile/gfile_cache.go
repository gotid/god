package gfile

import (
	"time"

	"github.com/gotid/god/internal/cmdenv"
	"github.com/gotid/god/lib/os/gcache"
	"github.com/gotid/god/lib/os/gfsnotify"
)

const (
	// Default expire time for file content caching in seconds.
	gDEFAULT_CACHE_EXPIRE = time.Minute
)

var (
	// Default expire time for file content caching.
	cacheExpire = cmdenv.Get("gf.gfile.cache", gDEFAULT_CACHE_EXPIRE).Duration()

	// internalCache is the memory cache for internal usage.
	internalCache = gcache.New()
)

// GetContents returns string content of given file by <path> from cache.
// If there's no content in the cache, it will read it from disk file specified by <path>.
// The parameter <expire> specifies the caching time for this file content in seconds.
func GetContentsWithCache(path string, duration ...time.Duration) string {
	return string(GetBytesWithCache(path, duration...))
}

// GetBinContents returns []byte content of given file by <path> from cache.
// If there's no content in the cache, it will read it from disk file specified by <path>.
// The parameter <expire> specifies the caching time for this file content in seconds.
func GetBytesWithCache(path string, duration ...time.Duration) []byte {
	key := cacheKey(path)
	expire := cacheExpire
	if len(duration) > 0 {
		expire = duration[0]
	}
	r, _ := internalCache.GetOrSetFuncLock(key, func() (interface{}, error) {
		b := GetBytes(path)
		if b != nil {
			// Adding this <path> to gfsnotify,
			// it will clear its cache if there's any changes of the file.
			_, _ = gfsnotify.Add(path, func(event *gfsnotify.Event) {
				internalCache.Remove(key)
				gfsnotify.Exit()
			})
		}
		return b, nil
	}, expire)
	if r != nil {
		return r.([]byte)
	}
	return nil
}

// cacheKey produces the cache key for gcache.
func cacheKey(path string) string {
	return "gf.gfile.cache:" + path
}
