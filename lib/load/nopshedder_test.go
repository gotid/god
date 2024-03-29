package load

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNopShedder(t *testing.T) {
	Disable()
	shedder := NewAdaptiveShedder()
	for i := 0; i < 1000; i++ {
		p, err := shedder.Allow()
		assert.Nil(t, err)
		p.Fail()
	}

	p, err := shedder.Allow()
	assert.Nil(t, err)
	p.Pass()
}
