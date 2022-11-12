package trace

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

const messageEvent = "message"

var (
	// MessageSent 是指发送消息的类型。
	MessageSent = messageType(RPCMessageTypeSent)
	// MessageReceived 是指接收消息的类型。
	MessageReceived = messageType(RPCMessageTypeReceived)
)

type messageType attribute.KeyValue

// Event 添加一个指定上下文、ID和消息到事件。
func (m messageType) Event(ctx context.Context, id int, message interface{}) {
	span := trace.SpanFromContext(ctx)
	if p, ok := message.(proto.Message); ok {
		span.AddEvent(messageEvent, trace.WithAttributes(
			attribute.KeyValue(m),
			RpcMessageIDKey.Int(id),
			RpcMessageUncompressedSizeKey.Int(proto.Size(p)),
		))
	} else {
		span.AddEvent(messageEvent, trace.WithAttributes(
			attribute.KeyValue(m),
			RpcMessageIDKey.Int(id),
		))
	}
}
