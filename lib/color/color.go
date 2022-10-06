package color

import "github.com/fatih/color"

type Color uint32

const (
	// NoColor 表示既无前景的业务背景色。
	NoColor Color = iota

	// 	FgBlack 表示前景色为黑色。
	FgBlack
	// 	FgWhite 表示前景色为白色。
	FgWhite
	// FgRed 表示前景色为红色。
	FgRed
	// FgGreen 表示前景色为绿色。
	FgGreen
	// 	FgBlue 表示前景色为蓝色。
	FgBlue
	// 	FgMagenta 表示前景色为品红色。
	FgMagenta
	// 	FgYellow 表示前景色为黄色。
	FgYellow
	// 	FgCyan 表示前景色为青色。
	FgCyan

	// 	BgBlack 表示背景色为黑色。
	BgBlack
	// 	BgWhite 表示背景色为白色。
	BgWhite
	// BgRed 表示背景色为红色。
	BgRed
	// BgGreen 表示背景色为绿色。
	BgGreen
	// 	BgBlue 表示背景色为蓝色。
	BgBlue
	// 	BgMagenta 表示背景色为品红色。
	BgMagenta
	// 	BgYellow 表示背景色为黄色。
	BgYellow
	// 	BgCyan 表示背景色为青色。
	BgCyan
)

var colors = map[Color][]color.Attribute{
	FgBlack:   {color.FgBlack, color.Bold},
	FgWhite:   {color.FgWhite, color.Bold},
	FgRed:     {color.FgRed, color.Bold},
	FgGreen:   {color.FgGreen, color.Bold},
	FgBlue:    {color.FgBlue, color.Bold},
	FgMagenta: {color.FgMagenta, color.Bold},
	FgYellow:  {color.FgYellow, color.Bold},
	FgCyan:    {color.FgCyan, color.Bold},

	BgBlack:   {color.BgBlack, color.Bold},
	BgWhite:   {color.BgWhite, color.Bold},
	BgRed:     {color.BgRed, color.Bold},
	BgGreen:   {color.BgGreen, color.Bold},
	BgBlue:    {color.BgBlue, color.Bold},
	BgMagenta: {color.BgMagenta, color.Bold},
	BgYellow:  {color.BgYellow, color.Bold},
	BgCyan:    {color.BgCyan, color.Bold},
}

// WithColor 返回应用了给定颜色的字符串。
func WithColor(text string, colour Color) string {
	c := color.New(colors[colour]...)
	return c.Sprint(text)
}

// WithColorPadding 返回应用了给定颜色的字符串，并带有前导和尾随空格。
func WithColorPadding(text string, colour Color) string {
	return WithColor(" "+text+" ", colour)
}
