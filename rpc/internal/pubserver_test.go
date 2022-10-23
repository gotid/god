package internal

import (
	"github.com/gotid/god/lib/netx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFigureOutListenOn(t *testing.T) {
	tests := []struct {
		input    string
		excepted string
	}{
		{
			input:    "192.168.0.5:1234",
			excepted: "192.168.0.5:1234",
		},
		{
			input:    "0.0.0.0:8080",
			excepted: netx.InternalIp() + ":8080",
		},
		{
			input:    ":8080",
			excepted: netx.InternalIp() + ":8080",
		},
	}

	for _, test := range tests {
		actual := figureOutListenOn(test.input)
		assert.Equal(t, test.excepted, actual)
	}
}
