package load

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSheddingStat(t *testing.T) {
	st := NewSheddingStat("foo")
	for i := 0; i < 3; i++ {
		st.IncrTotal()
	}
	for i := 0; i < 5; i++ {
		st.IncrPass()
	}
	for i := 0; i < 7; i++ {
		st.IncrDrop()
	}
	result := st.reset()
	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, int64(5), result.Pass)
	assert.Equal(t, int64(7), result.Drop)
}

func TestLoopTrue(t *testing.T) {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	close(ch)
	st := new(SheddingStat)
	st.name = "test"
	logEnabled.Set(true)
	st.loop(ch)
}

func TestLoopTrueAndDrop(t *testing.T) {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	close(ch)
	st := new(SheddingStat)
	st.IncrDrop()
	logEnabled.Set(true)
	st.loop(ch)
}

func TestLoopFalseAndDrop(t *testing.T) {
	ch := make(chan time.Time, 1)
	ch <- time.Now()
	close(ch)
	st := new(SheddingStat)
	st.IncrDrop()
	logEnabled.Set(false)
	st.loop(ch)
}
