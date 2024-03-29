package iox

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCountLines(t *testing.T) {
	const val = `1
2
3
4`
	file, err := os.CreateTemp(os.TempDir(), "test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	file.WriteString(val)
	file.Close()

	lines, err := CountLines(file.Name())
	assert.Nil(t, err)
	assert.Equal(t, 4, lines)
}
