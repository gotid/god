package stringx

import (
	"errors"
	"github.com/gotid/god/lib/lang"
)

var (
	// ErrInvalidStartPosition 表示起始位置无效。
	ErrInvalidStartPosition = errors.New("起始位置无效")
	// ErrInvalidStopPosition 表示结束位置无效。
	ErrInvalidStopPosition = errors.New("结束位置无效")
)

// Join 使用给定的分隔符连接任意数量的元素为一个字符串。
func Join(sep byte, elem ...string) string {
	var size int
	for _, e := range elem {
		size += len(e)
	}
	if size == 0 {
		return ""
	}

	buf := make([]byte, 0, size+len(elem)-1)
	for _, e := range elem {
		if len(e) == 0 {
			continue
		}

		if len(buf) > 0 {
			buf = append(buf, sep)
		}
		buf = append(buf, e...)
	}

	return string(buf)
}

// Contains 判断 list 中是否包含 str。
func Contains(list []string, str string) bool {
	for _, each := range list {
		if each == str {
			return true
		}
	}

	return false
}

// Filter 使用给定的过滤函数 filter 过滤字符串 s。
func Filter(s string, filter func(r rune) bool) string {
	var n int
	chars := []rune(s)
	for i, c := range chars {
		if n < i {
			chars[n] = c
		}
		if !filter(c) {
			n++
		}
	}

	return string(chars[:n])
}

// FirstN 从字符串 s 中返回前 n 个符文。
func FirstN(s string, n int, ellipsis ...string) string {
	var i int

	for j := range s {
		if i == n {
			ret := s[:j]
			for _, each := range ellipsis {
				ret += each
			}
			return ret
		}
		i++
	}

	return s
}

// HasEmpty 判断可变长度参数 args 中是否有空白字符串。
func HasEmpty(args ...string) bool {
	for _, arg := range args {
		if len(arg) == 0 {
			return true
		}
	}

	return false
}

// NotEmpty 判断可变长度参数 args 中是否无空白字符串。
func NotEmpty(args ...string) bool {
	return !HasEmpty(args...)
}

// Remove 从字符串 strings 中移除 ss。
func Remove(strings []string, ss ...string) []string {
	out := append([]string(nil), strings...)

	for _, s := range ss {
		var n int
		for _, v := range out {
			if v != s {
				out[n] = v
				n++
			}
		}
		out = out[:n]
	}

	return out
}

// Reverse 翻转字符串 s。
func Reverse(s string) string {
	runes := []rune(s)
	length := len(runes)

	for from, to := 0, length; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}

// Substr 返回 [start, stop) 之间的字符串。
// 无论字符串是 ascii 还是 utf8。
func Substr(s string, start, stop int) (string, error) {
	runes := []rune(s)
	length := len(runes)

	if start < 0 || start > length {
		return "", ErrInvalidStartPosition
	}
	if stop < 0 || stop > length {
		return "", ErrInvalidStopPosition
	}

	return string(runes[start:stop]), nil
}

// TakeOne 返回非空字符串 v 或其他字符串 or。
func TakeOne(v, or string) string {
	if len(v) > 0 {
		return v
	}

	return or
}

// TakeWithPriority 返回 fns 中第一个非空字符串。
func TakeWithPriority(fns ...func() string) string {
	for _, fn := range fns {
		val := fn()
		if len(val) > 0 {
			return val
		}
	}

	return ""
}

// Union 合并且返回两个字符串切片。
func Union(first, second []string) []string {
	set := make(map[string]lang.PlaceholderType)

	for _, each := range first {
		set[each] = lang.Placeholder
	}
	for _, each := range second {
		set[each] = lang.Placeholder
	}

	merged := make([]string, 0, len(set))
	for k := range set {
		merged = append(merged, k)
	}

	return merged
}
