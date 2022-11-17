package stringx

import "strings"

type (
	// Replacer 接口包装了 Replace 方法。
	Replacer interface {
		// Replace 使用给定的字典替换 text。
		Replace(text string) string
	}

	replacer struct {
		*node
		mapping map[string]string
	}
)

// NewReplacer 返回一个 Replacer。
func NewReplacer(mapping map[string]string) Replacer {
	rep := &replacer{
		node:    new(node),
		mapping: mapping,
	}
	for keyword := range mapping {
		rep.add(keyword)
	}
	rep.build()

	return rep
}

func (r *replacer) Replace(text string) string {
	var builder strings.Builder
	var start int
	chars := []rune(text)
	size := len(chars)

	for start < size {
		cur := r.node

		if start > 0 {
			builder.WriteString(string(chars[:start]))
		}

		for i := start; i < size; i++ {
			child, ok := cur.children[chars[i]]
			if ok {
				cur = child
			} else if cur == r.node {
				builder.WriteRune(chars[i])
				// cur 已指向 root，仅设置 start
				start = i + 1
				continue
			} else {
				curDepth := cur.depth
				cur = cur.fail
				child, ok = cur.children[chars[i]]
				if !ok {
					// 写入该路径
					builder.WriteString(string(chars[i-curDepth : i+1]))
					// 转到 root
					cur = r.node
					start = i + 1
					continue
				}

				failDepth := cur.depth
				// 跳转前写入路径
				builder.WriteString(string(chars[start : start+curDepth-failDepth]))
				start += curDepth - failDepth
				cur = child
			}

			if cur.end {
				val := string(chars[i+1-cur.depth : i+1])
				builder.WriteString(r.mapping[val])
				builder.WriteString(string(chars[i+1:]))
				// 仅匹配此路径，之前的所有路径都已完成
				if start >= i+1-cur.depth && i+1 >= size {
					return builder.String()
				}

				chars = []rune(builder.String())
				size = len(chars)
				builder.Reset()
				break
			}
		}

		if !cur.end {
			builder.WriteString(string(chars[start:]))
			return builder.String()
		}
	}

	return string(chars)
}
