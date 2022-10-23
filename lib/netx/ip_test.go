package netx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInternalIp(t *testing.T) {
	ip := InternalIp()
	fmt.Println(ip)

	assert.True(t, len(ip) > 0)
}
