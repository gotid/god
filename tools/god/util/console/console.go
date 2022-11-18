package console

import (
	"fmt"
	"github.com/gotid/god/tools/god/vars"
	"github.com/logrusorgru/aurora"
	"runtime"
)

type (
	// Console 控制台接口包装了 fmt.Sprintf。
	// colorConsole 向控制台提供彩色输出
	// ideaConsole 使用 intellij 的前缀输出
	Console interface {
		Info(format string, a ...any)
		Warning(format string, a ...any)
		Error(format string, a ...any)
		Success(format string, a ...any)
		Debug(format string, a ...any)
		MarkDone()
	}

	colorConsole struct {
		enable bool
	}

	// 用于 intellij 日志
	ideaConsole struct{}
)

func (i *ideaConsole) Info(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
}

func (i *ideaConsole) Warning(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println("[警告]：", msg)
}

func (i *ideaConsole) Error(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println("[错误]：", msg)
}

func (i *ideaConsole) Success(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println("[成功]：", msg)
}

func (i *ideaConsole) Debug(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(aurora.BrightCyan(msg))
}

func (i *ideaConsole) MarkDone() {
	i.Success("完成。")
}

func (c *colorConsole) Info(format string, a ...any) {
	if !c.enable {
		return
	}

	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
}

func (c *colorConsole) Warning(format string, a ...any) {
	if !c.enable {
		return
	}
	msg := fmt.Sprintf(format, a...)
	println(aurora.BrightYellow(msg))
}

func (c *colorConsole) Error(format string, a ...any) {
	if !c.enable {
		return
	}
	msg := fmt.Sprintf(format, a...)
	println(aurora.BrightRed(msg))
}

func (c *colorConsole) Success(format string, a ...any) {
	if !c.enable {
		return
	}
	msg := fmt.Sprintf(format, a...)
	println(aurora.BrightGreen(msg))
}

func (c *colorConsole) Debug(format string, a ...any) {
	if !c.enable {
		return
	}
	msg := fmt.Sprintf(format, a...)
	println(aurora.BrightCyan(msg))
}

func (c *colorConsole) MarkDone() {
	if !c.enable {
		return
	}

	c.Success("完成。")
}

// NewConsole 返回一个 Console 控制台。
func NewConsole(idea bool) Console {
	if idea {
		return NewIdeaConsole()
	}

	return NewColorConsole()
}

// NewIdeaConsole 返回一个 intellij 控制台。
func NewIdeaConsole() Console {
	return &ideaConsole{}
}

// NewColorConsole 返回一个彩色输出的控制台。
func NewColorConsole(enable ...bool) Console {
	logEnable := true
	for _, e := range enable {
		logEnable = e
	}

	return &colorConsole{
		enable: logEnable,
	}
}

var defaultConsole = &colorConsole{enable: true}

func println(msg any) {
	value, ok := msg.(aurora.Value)
	if !ok {
		fmt.Println(msg)
	}

	goos := runtime.GOOS
	if goos == vars.OsWindows {
		fmt.Println(value.Value())
		return
	}

	fmt.Println(msg)
}

func Info(format string, a ...any) {
	defaultConsole.Info(format, a...)
}

func Warning(format string, a ...any) {
	defaultConsole.Warning(format, a...)
}

func Error(format string, a ...any) {
	defaultConsole.Error(format, a...)
}
