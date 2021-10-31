package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const sleepInterval = 10 * time.Millisecond

func TestElapsedTimer_Duration(t *testing.T) {
	timer := NewElapsedTimer()
	time.Sleep(sleepInterval)
	assert.True(t, timer.Duration() >= sleepInterval)
}

func TestElapsedTimer_Elapsed(t *testing.T) {
	timer := NewElapsedTimer()
	time.Sleep(sleepInterval)
	duration, err := time.ParseDuration(timer.Elapsed())
	assert.Nil(t, err)
	assert.True(t, duration >= sleepInterval)
}

func TestElapsedTimer_ElapsedMs(t *testing.T) {
	timer := NewElapsedTimer()
	time.Sleep(sleepInterval)
	duration, err := time.ParseDuration(timer.ElapsedMs())
	assert.Nil(t, err)
	assert.True(t, duration >= sleepInterval)
}

func TestCurrent(t *testing.T) {
	currentMillis := CurrentMillis()
	currentMicros := CurrentMicros()
	fmt.Println(currentMillis)
	fmt.Println(currentMicros)
	assert.True(t, currentMillis > 0)
	assert.True(t, currentMicros > 0)
	assert.True(t, currentMillis*1000 <= currentMicros)
}
