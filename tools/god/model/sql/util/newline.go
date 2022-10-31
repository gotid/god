package util

import "strings"

// TrimNewLine 移除换行符 \r 和 \n。
func TrimNewLine(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}
