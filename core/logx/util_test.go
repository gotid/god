package logx

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCaller(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	assert.Contains(t, getCaller(1), filepath.Base(file))
	assert.True(t, len(getCaller(1<<10)) == 0)
}

func TestGetTimestamp(t *testing.T) {
	ts := getTimestamp()
	tm, err := time.Parse(timeFormat, ts)
	assert.Nil(t, err)
	assert.True(t, time.Since(tm) < time.Minute)
}

func TestPrettyCaller(t *testing.T) {
	tests := []struct {
		name string
		file string
		line int
		want string
	}{
		{
			name: "常规文件路径",
			file: "util_test.go",
			line: 123,
			want: "util_test.go:123",
		},
		{
			name: "相对文件路径",
			file: "logx/util_test.go",
			line: 123,
			want: "logx/util_test.go:123",
		},
		{
			name: "较长文件路径",
			file: "github.com/gotid/god/core/logx/util_test.go",
			line: 12,
			want: "logx/util_test.go:12",
		},
		{
			name: "本地文件路径",
			file: "/Users/zs/god/core/logx/util_test.go",
			line: 1234,
			want: "logx/util_test.go:1234",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, prettyCaller(test.file, test.line))
		})
	}
}

func BenchmarkGetCaller(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		getCaller(i)
	}
}
