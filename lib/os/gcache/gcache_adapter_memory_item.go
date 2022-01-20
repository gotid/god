package gcache

import (
	"github.com/gotid/god/lib/os/gtime"
)

// IsExpired checks whether <item> is expired.
func (item *adapterMemoryItem) IsExpired() bool {
	// Note that it should use greater than or equal judgement here
	// imagining that the cache time is only 1 millisecond.
	if item.e >= gtime.TimestampMilli() {
		return false
	}
	return true
}
