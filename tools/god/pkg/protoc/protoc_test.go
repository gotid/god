package protoc

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion(t *testing.T) {
	version, err := Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, version)
	fmt.Println(version)
}
