package logx

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithDurationError(t *testing.T) {
	var b strings.Builder
	log.SetOutput(&b)
	WithDuration(time.Second).Error("foo")
	assert.True(t, strings.Contains(b.String(), "duration"), b.String())
	fmt.Println(b.String())
}

func TestWithDurationErrorf(t *testing.T) {
	var b strings.Builder
	log.SetOutput(&b)
	WithDuration(time.Second).Errorf("foo")
	assert.True(t, strings.Contains(b.String(), "duration"), b.String())
	fmt.Println(b.String())
}

func TestWithDurationInfo(t *testing.T) {
	var b strings.Builder
	log.SetOutput(&b)
	WithDuration(time.Second).Info("foo")
	assert.True(t, strings.Contains(b.String(), "duration"), b.String())
	fmt.Println(b.String())
}

func TestWithDurationInfof(t *testing.T) {
	var b strings.Builder
	log.SetOutput(&b)
	WithDuration(time.Second).Infof("foo")
	assert.True(t, strings.Contains(b.String(), "duration"), b.String())
	fmt.Println(b.String())
}

func TestWithDurationSlow(t *testing.T) {
	var b strings.Builder
	log.SetOutput(&b)
	WithDuration(time.Second).Slow("foo")
	assert.True(t, strings.Contains(b.String(), "duration"), b.String())
	fmt.Println(b.String())
}

func TestWithDurationSlowf(t *testing.T) {
	var b strings.Builder
	log.SetOutput(&b)
	WithDuration(time.Second).WithDuration(time.Hour).Slowf("foo")
	assert.True(t, strings.Contains(b.String(), "duration"), b.String())
	fmt.Println(b.String())
}
