//go:build linux
// +build linux

package stat

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
)

func TestReport(t *testing.T) {
	os.Setenv(clusterNameKey, "test-cluster")
	defer os.Unsetenv(clusterNameKey)

	var count int32
	SetReporter(func(s string) {
		atomic.AddInt32(&count, 1)
	})
	for i := 0; i < 10; i++ {
		Report(strconv.Itoa(i))
	}
	assert.Equal(t, int32(1), count)
}
