package proc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDone(t *testing.T) {
	select {
	case <-Done():
		assert.Fail(t, "应当运行")
	default:
		assert.NotNil(t, Done())
	}
}
