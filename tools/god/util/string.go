package util

import (
	"github.com/gotid/god/tools/god/util/console"
	"strings"
)

var goKeyword = map[string]string{
	"var":         "variable",
	"const":       "constant",
	"package":     "pkg",
	"func":        "function",
	"return":      "rtn",
	"defer":       "dfr",
	"go":          "goo",
	"select":      "slt",
	"struct":      "structure",
	"interface":   "itf",
	"chan":        "channel",
	"type":        "tp",
	"map":         "mp",
	"range":       "rg",
	"break":       "brk",
	"case":        "caz",
	"continue":    "ctn",
	"for":         "fr",
	"fallthrough": "fth",
	"else":        "es",
	"if":          "ef",
	"switch":      "swt",
	"goto":        "gt",
	"default":     "dft",
}

// SafeString 将给定字符串转为 golang 中安全命名样式的字符串。
// 将非字母、非数字转为下划线。
func SafeString(s string) string {
	if len(s) == 0 {
		return s
	}

	data := strings.Map(func(r rune) rune {
		if isSafeRune(r) {
			return r
		}
		return '_'
	}, s)

	headRune := rune(data[0])
	if isNumber(headRune) {
		return "_" + data
	}

	return data
}

func isSafeRune(r rune) bool {
	return isLetter(r) || isNumber(r) || r == '_'
}

func isNumber(r rune) bool {
	return '0' <= r && r <= '9'
}

func isLetter(r rune) bool {
	return 'A' <= r && r <= 'z'
}

// EscapeGolangKeyword 转义 golang 关键字。
func EscapeGolangKeyword(s string) string {
	if !isGolangKeywords(s) {
		return s
	}

	r := goKeyword[s]
	console.Info("[EscapeGolangKeyword]：go 关键字 %q 禁止使用，已转为 %q", s, r)

	return r
}

func isGolangKeywords(s string) bool {
	_, ok := goKeyword[s]
	return ok
}

// Title 转为首字母大写。
func Title(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

// UnTitle 取消首字母大写。
func UnTitle(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToLower(s[:1]) + s[1:]
}

// Index 返回切片中给定 item 的索引，如未找到返回 -1。
func Index(slice []string, item string) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}

	return -1
}
