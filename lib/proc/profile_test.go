package proc

import (
	"fmt"
	"github.com/gotid/god/lib/logx"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestProfile(t *testing.T) {
	var buf strings.Builder
	w := logx.NewWriter(&buf)
	o := logx.Reset()
	logx.SetWriter(w)

	defer func() {
		logx.Reset()
		logx.SetWriter(o)
	}()

	profiler := StartProfile()
	// 不能再次启动
	assert.NotNil(t, StartProfile())
	profiler.Stop()
	// 再次关闭
	profiler.Stop()
	assert.True(t, strings.Contains(buf.String(), ".pprof"))
	fmt.Println(buf.String())
}
