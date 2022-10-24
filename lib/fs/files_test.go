package fs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCloseOnExec(t *testing.T) {
	file := os.NewFile(0, os.DevNull)
	assert.NotPanics(t, func() {
		CloseOnExec(file)
	})
}
