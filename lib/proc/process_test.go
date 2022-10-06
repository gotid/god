package proc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessName(t *testing.T) {
	assert.True(t, len(ProcessName()) > 0)
}

func TestPid(t *testing.T) {
	assert.True(t, Pid() > 0)
}
