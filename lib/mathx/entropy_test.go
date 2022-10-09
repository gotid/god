package mathx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalcEntropy(t *testing.T) {
	const total, count = 1000, 100
	m := make(map[interface{}]int, total)
	for i := 0; i < total; i++ {
		m[i] = count
	}
	assert.True(t, CalcEntropy(m) > .99)
}
