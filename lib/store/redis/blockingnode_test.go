package redis

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateBlockingNode(t *testing.T) {
	r, err := miniredis.Run()
	assert.Nil(t, err)
	node, err := CreateBlockingNode(New(r.Addr()))
	assert.Nil(t, err)
	node.Close()

	node, err = CreateBlockingNode(New(r.Addr(), WithCluster()))
	assert.Nil(t, err)
	node.Close()
}
