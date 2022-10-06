//go:build linux || darwin
// +build linux darwin

package proc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestShutdown(t *testing.T) {
	SetTimeToForceQuit(time.Hour)
	assert.Equal(t, time.Hour, delayTimeBeforeForceQuit)

	var val int
	called := AddWrapUpListener(func() {
		val++
	})
	wrapUpListeners.notifyListeners()
	called()
	assert.Equal(t, 1, val)

	called = AddShutdownListener(func() {
		val += 2
	})
	shutdownListeners.notifyListeners()
	called()
	assert.Equal(t, 3, val)
}
