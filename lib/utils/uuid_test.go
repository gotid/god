package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUUID(t *testing.T) {
	uuid := NewUUID()
	assert.Equal(t, 36, len(uuid))
	fmt.Println(uuid)
}
