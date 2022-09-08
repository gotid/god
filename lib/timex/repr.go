package timex

import (
	"fmt"
	"time"
)

// ReprOfDuration 返回给定毫秒时长的字符串形式。
func ReprOfDuration(duration time.Duration) string {
	return fmt.Sprintf("%.1fms", float32(duration)/float32(time.Millisecond))
}
