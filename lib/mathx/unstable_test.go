package mathx

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestUnstable_AroundDuration(t *testing.T) {
	u := NewUnstable(0.05)
	for i := 0; i < 1000; i++ {
		v := u.AroundDuration(time.Second)
		assert.True(t, float64(time.Second)*0.95 <= float64(v))
		assert.True(t, float64(v) <= float64(time.Second)*1.05)
	}
}

func TestUnstable_AroundInt(t *testing.T) {
	const target = 10000
	u := NewUnstable(0.05)
	for i := 0; i < 1000; i++ {
		v := u.AroundInt(target)
		assert.True(t, float64(target)*0.95 <= float64(v))
		assert.True(t, float64(v) <= float64(target)*1.05)
	}
}

func TestUnstable_AroundIntLarge(t *testing.T) {
	const target int64 = 10000
	unstable := NewUnstable(5)
	for i := 0; i < 1000; i++ {
		val := unstable.AroundInt(target)
		assert.True(t, 0 <= val)
		assert.True(t, val <= 2*target)
	}
}

func TestUnstable_AroundIntNegative(t *testing.T) {
	const target int64 = 10000
	unstable := NewUnstable(-0.05)
	for i := 0; i < 1000; i++ {
		val := unstable.AroundInt(target)
		assert.Equal(t, target, val)
	}
}

func TestUnstable_Distribution(t *testing.T) {
	const (
		seconds = 10000
		total   = 10000
	)

	m := make(map[int]int)
	expire := NewUnstable(0.05)
	for i := 0; i < total; i++ {
		val := int(expire.AroundInt(seconds))
		m[val]++
	}

	_, ok := m[0]
	assert.False(t, ok)

	mi := make(map[interface{}]int, len(m))
	for k, v := range m {
		mi[k] = v
	}
	entropy := CalcEntropy(mi)
	assert.True(t, len(m) > 1)
	assert.True(t, entropy > 0.95)
}
