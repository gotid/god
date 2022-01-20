package breaker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/timex"
)

const (
	numHistoryReasons = 5
	timeFormat        = "15:04:05"
)

// errorWindow 一扇观察最近5条记录的信息窗户
type errorWindow struct {
	reasons [numHistoryReasons]string
	index   int
	count   int
	lock    sync.Mutex
}

func (ew *errorWindow) add(reason string) {
	ew.lock.Lock()
	ew.reasons[ew.index] = fmt.Sprintf("%s %s", timex.Time().Format(timeFormat), reason)
	ew.index = (ew.index + 1) % numHistoryReasons
	ew.count = mathx.MinInt(ew.count+1, numHistoryReasons)
	ew.lock.Unlock()
}

// String 按时间降序输出
func (ew *errorWindow) String() string {
	var reasons []string

	ew.lock.Lock()
	// 反向排序
	for i := ew.index - 1; i >= ew.index-ew.count; i-- {
		reasons = append(reasons, ew.reasons[(i+numHistoryReasons)%numHistoryReasons])
	}
	ew.lock.Unlock()

	return strings.Join(reasons, "\n")
}
