package logx

import (
	"sync/atomic"
	"testing"

	"github.com/gotid/god/lib/color"
	"github.com/stretchr/testify/assert"
)

func TestWithColor(t *testing.T) {
	old := atomic.SwapUint32(&encoding, plainEncodingType)
	defer atomic.StoreUint32(&encoding, old)

	output := WithColor("hellox", color.BgBlue)
	assert.Equal(t, "hellox", output)

	atomic.StoreUint32(&encoding, jsonEncodingType)
	output = WithColor("hellox", color.BgBlue)
	assert.Equal(t, "hellox", output)
}

func TestWithColorPadding(t *testing.T) {
	old := atomic.SwapUint32(&encoding, plainEncodingType)
	defer atomic.StoreUint32(&encoding, old)

	output := WithColorPadding("hellox", color.BgBlue)
	assert.Equal(t, " hellox ", output)

	atomic.StoreUint32(&encoding, jsonEncodingType)
	output = WithColorPadding("hellox", color.BgBlue)
	assert.Equal(t, "hellox", output)
}
