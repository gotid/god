package resolver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildDirectTarget(t *testing.T) {
	target := BuildDirectTarget([]string{"localhost:123", "localhost:456"})
	assert.Equal(t, "direct:///localhost:123,localhost:456", target)
}

func TestBuildDiscovTarget(t *testing.T) {
	target := BuildDiscovTarget([]string{"localhost:123", "localhost:456"}, "foo")
	assert.Equal(t, "etcd://localhost:123,localhost:456/foo", target)
}
