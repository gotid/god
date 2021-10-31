package load

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewShedderGroup(t *testing.T) {
	group := NewShedderGroup()
	t.Run("get", func(t *testing.T) {
		shedder := group.GetShedder("test")
		assert.NotNil(t, shedder)
	})
}

func TestShedderClose(t *testing.T) {
	var nop nopCloser
	assert.Nil(t, nop.Close())
}
