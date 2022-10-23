package resolver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegister(t *testing.T) {
	assert.NotPanics(t, func() {
		Register()
	})
}
