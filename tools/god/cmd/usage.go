package cmd

import (
	"fmt"
	"github.com/gotid/god/tools/god/vars"
	"github.com/logrusorgru/aurora"
	"runtime"
)

var colorRender = []func(v interface{}) string{
	func(v interface{}) string {
		return aurora.BrightRed(v).String()
	},
	func(v interface{}) string {
		return aurora.BrightGreen(v).String()
	},
	func(v interface{}) string {
		return aurora.BrightYellow(v).String()
	},
	func(v interface{}) string {
		return aurora.BrightBlue(v).String()
	},
	func(v interface{}) string {
		return aurora.BrightMagenta(v).String()
	},
	func(v interface{}) string {
		return aurora.BrightCyan(v).String()
	},
}

func blue(s string) string {
	if runtime.GOOS == vars.OsWindows {
		return s
	}

	return aurora.BrightBlue(s).String()
}

func green(s string) string {
	if runtime.GOOS == vars.OsWindows {
		return s
	}

	return aurora.BrightGreen(s).String()
}

func rainbow(s string) string {
	if runtime.GOOS == vars.OsWindows {
		return s
	}

	s0 := s[0]
	return colorRender[int(s0)%(len(colorRender)-1)](s)
}

// 对字符串 s 进行右侧填充。
func rPadX(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return rainbow(fmt.Sprintf(template, s))
}
