package parser

import (
	"github.com/emicklei/proto"
	"strings"
)

// GetComment 返回以 // 为前缀的内容。
func GetComment(comment *proto.Comment) string {
	if comment == nil {
		return ""
	}

	return "// " + strings.TrimSpace(comment.Message())
}
