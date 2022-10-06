package mathx

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestProba_TrueOnProba(t *testing.T) {
	const (
		proba   = math.Pi / 10
		total   = 100000
		epsilon = 0.05
	)
	var count int
	p := NewProba()
	for i := 0; i < total; i++ {
		if p.TrueOnProba(proba) {
			count++
		}
	}

	ratio := float64(count) / float64(total)
	assert.InEpsilon(t, proba, ratio, epsilon)
}
