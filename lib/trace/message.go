package trace

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"go.opentelemetry.io/otel/attribute"
)

const messageEvent = "message"

var (
	// MessageSent 发送消息的类型。
	MessageSent = messageType(RPCMessageTypeSent)
	// MessageReceived 接收消息的类型。
	MessageReceived = messageType(RPCMessageTypeReceived)
)

type messageType attribute.KeyValue

// Event 将 messageType 事件添加到相关 span 行为。
func (t messageType) Event(ctx context.Context, id int, message interface{}) {
	span := trace.SpanFromContext(ctx)
	if p, ok := message.(proto.Message); ok {
		span.AddEvent(messageEvent, trace.WithAttributes(
			attribute.KeyValue(t),
			RPCMessageIDKey.Int(id),
			RPCMessageUncompressedSizeKey.Int(proto.Size(p)),
		))
	} else {
		span.AddEvent(messageEvent, trace.WithAttributes(
			attribute.KeyValue(t),
			RPCMessageIDKey.Int(id),
		))
	}
}
