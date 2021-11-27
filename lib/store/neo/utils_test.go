package neo

import (
	"testing"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/stretchr/testify/assert"
)

func TestLabels(t *testing.T) {
	n := neo4j.Node{
		Labels: nil,
	}
	assert.Equal(t, "", Labels(n))

	n = neo4j.Node{
		Labels: []string{},
	}
	assert.Equal(t, "", Labels(n))

	n = neo4j.Node{
		Labels: []string{"User"},
	}
	assert.Equal(t, "User", Labels(n))

	n = neo4j.Node{
		Labels: []string{"User", "Media"},
	}
	assert.Equal(t, "User:Media", Labels(n))
}
