package sqlx

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mapping"
	"strconv"
	"strings"
	"time"
)

var errUnbalancedEscape = errors.New("逃逸字符后面没有字符")

func logInstanceError(dsn string, err error) {
	dsn = desensitize(dsn)
	logx.Errorf("SQL 实例错误：%s - %v", dsn, err)
}

func logSQLError(ctx context.Context, stmt string, err error) {
	if err != nil && err != ErrNotFound {
		logx.WithContext(ctx).Errorf("SQL 语句：%s，错误：%s", stmt, err.Error())
	}
}

// 对数据库连接进行脱敏。
func desensitize(dsn string) string {
	// 移除账号
	pos := strings.LastIndex(dsn, "@")
	if 0 <= pos && pos+1 < len(dsn) {
		dsn = dsn[pos+1:]
	}

	return dsn
}

// 格式化查询语句。
func format(query string, args ...any) (string, error) {
	numArgs := len(args)
	if numArgs == 0 {
		return query, nil
	}

	var b strings.Builder
	var argIndex int
	queryLength := len(query)

	for i := 0; i < queryLength; i++ {
		ch := query[i]
		switch ch {
		case '?':
			if argIndex >= numArgs {
				return "", fmt.Errorf("错误：SQL 中有 %d 个 ?，但只提供了 %d 个参数值", argIndex+1, numArgs)
			}

			writeValue(&b, args[argIndex])
			argIndex++
		case ':', '$':
			var j int
			for j = i + 1; j < queryLength; j++ {
				char := query[j]
				if char < '0' || '9' < char {
					break
				}
			}

			if j > i+1 {
				index, err := strconv.Atoi(query[i+1 : j])
				if err != nil {
					return "", err
				}

				// pg 和 oracle 的索引从 1 开始
				if index > argIndex {
					argIndex = index
				}

				index--
				if index < 0 || numArgs <= index {
					return "", fmt.Errorf("错误：索引 %d 越界", index)
				}

				writeValue(&b, args[index])
				i = j - 1
			}
		case '\'', '"', '`':
			b.WriteByte(ch)

			for j := i + 1; j < queryLength; j++ {
				char := query[j]
				b.WriteByte(char)

				if char == '\\' {
					j++
					if j >= queryLength {
						return "", errUnbalancedEscape
					}

					b.WriteByte(query[j])
				} else if char == ch {
					i = j
					break
				}
			}
		default:
			b.WriteByte(ch)
		}
	}

	if argIndex < numArgs {
		return "", fmt.Errorf("错误：占位符(?)的个数为 %d，比提供的参数要少", argIndex)
	}

	return b.String(), nil
}

func writeValue(b *strings.Builder, arg any) {
	switch v := arg.(type) {
	case bool:
		if v {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	case string:
		b.WriteByte('\'')
		b.WriteString(escape(v))
		b.WriteByte('\'')
	case time.Time:
		b.WriteByte('\'')
		b.WriteString(v.String())
		b.WriteByte('\'')
	case *time.Time:
		b.WriteByte('\'')
		b.WriteString(v.String())
		b.WriteByte('\'')
	default:
		b.WriteString(mapping.Repr(v))
	}
}

func escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\x00':
			b.WriteString(`\x00`)
		case '\r':
			b.WriteString(`\r`)
		case '\n':
			b.WriteString(`\n`)
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '"':
			b.WriteString(`\"`)
		case '\x1a':
			b.WriteString(`\x1a`)
		default:
			b.WriteRune(r)
		}
	}

	return b.String()
}
