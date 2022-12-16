package stringx

import (
	"bytes"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
	"unicode"
)

var Whitespace = []rune{'\n', '\t', '\f', '\v', ' '}

type String struct {
	source string
}

// Source 返回源文本的字符串值。
func (s String) Source() string {
	return s.source
}

// Title 首字母大写。
func (s String) Title() string {
	if s.IsEmptyOrSpace() {
		return s.source
	}

	return cases.Title(language.English, cases.NoLower).String(s.source)
}

// UnTitle 首字母小写。
// 若0索引符文不是字母，则返回原始字符串。
func (s String) UnTitle() string {
	if s.IsEmptyOrSpace() {
		return s.source
	}

	r := rune(s.source[0])
	if !unicode.IsUpper(r) && !unicode.IsLower(r) {
		return s.source
	}

	return string(unicode.ToLower(r)) + s.source[1:]
}

// IsEmptyOrSpace 根据去除字符串两端空格后的长度，判断字符串是否为空。
func (s String) IsEmptyOrSpace() bool {
	if len(s.source) == 0 {
		return true
	}

	if strings.TrimSpace(s.source) == "" {
		return true
	}

	return false
}

// ToCamel 转为大驼峰式。
func (s String) ToCamel() string {
	list := s.splitBy(func(r rune) bool {
		return r == '_'
	}, true)
	var target []string
	for _, item := range list {
		target = append(target, From(item).Title())
	}

	return strings.Join(target, "")
}

// ToSnake 转为蛇式。
func (s String) ToSnake() string {
	list := s.splitBy(unicode.IsUpper, false)
	var target []string
	for _, item := range list {
		target = append(target, From(item).ToLower())
	}
	return strings.Join(target, "_")
}

// ToLower 转为小写。
func (s String) ToLower() string {
	return strings.ToLower(s.source)
}

// ToUpper 转为大写。
func (s String) ToUpper() string {
	return strings.ToUpper(s.source)
}

// 按给定函数进行字符串分割
func (s String) splitBy(fn func(r rune) bool, remove bool) []string {
	if s.IsEmptyOrSpace() {
		return nil
	}

	var list []string
	buffer := new(bytes.Buffer)
	for _, r := range s.source {
		if fn(r) {
			if buffer.Len() != 0 {
				list = append(list, buffer.String())
				buffer.Reset()
			}
			if !remove {
				buffer.WriteRune(r)
			}
			continue
		}
		buffer.WriteRune(r)
	}
	if buffer.Len() != 0 {
		list = append(list, buffer.String())
	}

	return list
}

// From 将输入文本转为 String 并返回。
func From(data string) String {
	return String{
		source: data,
	}
}

// ContainsWhitespace 判断字符串中是否包含空白字符串。
func ContainsWhitespace(s string) bool {
	return ContainsAny(s, Whitespace...)
}

// ContainsAny 判断字符串中是否包含给定的符文。
func ContainsAny(s string, runes ...rune) bool {
	if len(runes) == 0 {
		return true
	}

	tmp := make(map[rune]struct{}, len(runes))
	for _, r := range runes {
		tmp[r] = struct{}{}
	}

	for _, r := range s {
		if _, ok := tmp[r]; ok {
			return true
		}
	}

	return false
}
