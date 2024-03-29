package timex

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReprOfDuration(t *testing.T) {
	assert.Equal(t, "1000.0ms", ReprOfDuration(time.Second))
	assert.Equal(t, "1111.6ms", ReprOfDuration(time.Second+time.Millisecond*111+time.Microsecond*555))
}
