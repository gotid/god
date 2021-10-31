package load

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShedderStat(t *testing.T) {
	stat := NewShedderStat("any")
	for i := 0; i < 3; i++ {
		stat.IncrTotal()
	}
	for i := 0; i < 5; i++ {
		stat.IncrPass()
	}
	for i := 0; i < 7; i++ {
		stat.IncrDrop()
	}
	result := stat.reset()
	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, int64(5), result.Pass)
	assert.Equal(t, int64(7), result.Drop)
}

func TestLoopTrue(t *testing.T) {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	close(ch)

	stat := new(ShedderStat)
	logEnabled.Set(true)
	stat.loop(ch)
}

func TestLoopTrueAndDrop(t *testing.T) {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	close(ch)

	stat := new(ShedderStat)
	stat.IncrDrop()
	logEnabled.Set(true)
	stat.loop(ch)
}

func TestLoopFalseAndDrop(t *testing.T) {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	close(ch)

	stat := new(ShedderStat)
	stat.IncrDrop()
	logEnabled.Set(false)
	stat.loop(ch)
}
