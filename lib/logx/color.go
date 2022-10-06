package logx

import (
	"sync/atomic"

	"github.com/gotid/god/lib/color"
)

// WithColor 是一个助手函数，用于为字符串添加颜色，尽在纯文本编码模式下生效。
func WithColor(text string, colour color.Color) string {
	if atomic.LoadUint32(&encoding) == plainEncodingType {
		return color.WithColor(text, colour)
	}

	return text
}

// WithColorPadding 是一个助手函数，返回应用了给定颜色的字符串，并带有前导和尾随空格。
func WithColorPadding(text string, colour color.Color) string {
	if atomic.LoadUint32(&encoding) == plainEncodingType {
		return color.WithColorPadding(text, colour)
	}

	return text
}
