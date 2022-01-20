package logx

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/gotid/god/lib/timex"
	"github.com/stretchr/testify/assert"
)

func TestLimitedLogger(t *testing.T) {
	tests := []struct {
		name      string
		duration  time.Duration
		lastTime  time.Duration
		discarded uint32
		logged    bool
	}{
		{
			name:   "空记录器",
			logged: true,
		},
		{
			name:      "常规记录器",
			duration:  time.Hour,
			lastTime:  timex.Now(),
			discarded: 10,
			logged:    false,
		},
		{
			name:      "慢日志记录器",
			duration:  time.Duration(1),
			lastTime:  -1000,
			discarded: 10,
			logged:    true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			logger := newLimitedLogger(0)
			logger.duration = test.duration
			logger.discarded = test.discarded
			logger.lastTime.Set(test.lastTime)

			var run int32
			logger.logOrDiscard(func() {
				atomic.AddInt32(&run, 1)
			})
			if test.logged {
				assert.Equal(t, int32(1), atomic.LoadInt32(&run))
			} else {
				assert.Equal(t, int32(0), atomic.LoadInt32(&run))
				assert.Equal(t, test.discarded+1, atomic.LoadUint32(&logger.discarded))
			}
		})
	}
}
