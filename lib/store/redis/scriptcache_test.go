package redis

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScriptCache(t *testing.T) {
	cache := GetScriptCache()
	cache.SetSha("foo", "bar")
	cache.SetSha("bla", "blabla")

	bar, ok := cache.GetSha("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", bar)
}
