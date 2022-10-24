package parser

import "github.com/emicklei/proto"

// Message 内嵌 proto.Message。
type Message struct {
	*proto.Message
}
