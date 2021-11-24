package neo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	target   = "bolt://localhost:7687"
	username = "neo4j"
	password = "asdfasdf"
	limit    = 10
	timeout  = 50 * time.Millisecond
)

var cfg *Config

func init() {
	cfg = NewConfig(target, username, password, limit, timeout)
}

func TestConfig(t *testing.T) {
	assert.EqualValues(t, []interface{}{
		target,
		username,
		password,
		limit,
		timeout,
	}, []interface{}{
		cfg.Target,
		cfg.Username,
		cfg.Password,
		cfg.Limit,
		cfg.Timeout,
	})
}
