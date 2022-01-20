package glog

import "github.com/gotid/god/lib/container/gmap"

const (
	// Default group name for instance usage.
	DEFAULT_NAME = "default"
)

// Instances map.
var instances = gmap.NewStrAnyMap(true)

// Instance returns an instance of Logger with default settings.
// The parameter <name> is the name for the instance.
func Instance(name ...string) *Logger {
	key := DEFAULT_NAME
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	return instances.GetOrSetFuncLock(key, func() interface{} {
		return New()
	}).(*Logger)
}
