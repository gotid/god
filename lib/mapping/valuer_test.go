package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapValuerWithInherit_Value(t *testing.T) {
	input := map[string]any{
		"discovery": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
		"component": map[string]any{
			"name": "test",
		},
	}
	valuer := recursiveValuer{
		current: mapValuer(input["component"].(map[string]any)),
		parent: simpleValuer{
			current: mapValuer(input),
		},
	}

	val, ok := valuer.Value("discovery")
	assert.True(t, ok)

	m, ok := val.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "localhost", m["host"])
	assert.Equal(t, 8080, m["port"])
}
