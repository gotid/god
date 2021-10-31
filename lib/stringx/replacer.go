package stringx

import "strings"

type (
	// Replacer 是一个包装替换函数的接口。
	Replacer interface {
		Replace(s string) string
	}

	replacer struct {
		node
		mapping map[string]string
	}
)

func NewReplacer(m map[string]string) Replacer {
	r := &replacer{mapping: m}
	for k := range m {
		r.add(k)
	}
	return r
}

func (r replacer) Replace(s string) string {
	var b strings.Builder
	chars := []rune(s)
	size := len(chars)
	start := -1

	for i := 0; i < size; i++ {
		child, ok := r.children[chars[i]]
		if !ok {
			b.WriteRune(chars[i])
			continue
		}

		if start < 0 {
			start = i
		}
		end := -1
		if child.end {
			end = i + 1
		}

		j := i + 1
		for ; j < size; j++ {
			grandchild, ok := child.children[chars[j]]
			if !ok {
				break
			}

			child = grandchild
			if child.end {
				end = j + 1
				i = j
			}
		}

		if end > 0 {
			i = j - 1
			b.WriteString(r.mapping[string(chars[start:end])])
		} else {
			b.WriteRune(chars[i])
		}
		start = -1
	}

	return b.String()
}
