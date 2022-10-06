package trace

import (
	"github.com/stretchr/testify/assert"
	gcodes "google.golang.org/grpc/codes"
	"testing"
)

func TestStatusCodeAttr(t *testing.T) {
	assert.Equal(t, GRPCStatusCodeKey.Int(int(gcodes.DataLoss)), StatusCodeAttr(gcodes.DataLoss))
}
