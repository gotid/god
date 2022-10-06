package trace

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	gcodes "google.golang.org/grpc/codes"
)

const (
	// GRPCStatusCodeKey 是 GRPC请求的数字状态码的约定方式。
	GRPCStatusCodeKey = attribute.Key("rpc.grpc.status_code")
	// RPCNameKey 是传输或接收的消息名称。
	RPCNameKey = attribute.Key("name")
	// RPCMessageTypeKey 是传输或接收的消息类型。
	RPCMessageTypeKey = attribute.Key("message.type")
	// RPCMessageIDKey 是传输或接收的消息标识符。
	RPCMessageIDKey = attribute.Key("message.id")
	// RPCMessageCompressedSizeKey 是传输或接收的消息压缩后的字节大小。
	RPCMessageCompressedSizeKey = attribute.Key("message.compressed_size")
	// RPCMessageUncompressedSizeKey 是传输或接收的消息未压缩的字节大小。
	RPCMessageUncompressedSizeKey = attribute.Key("message.uncompressed_size")
)

// 常见 RPC 属性的语义约定。
var (
	// RPCSystemGRPC 是GRPC作为远程处理系统的语义约定。
	RPCSystemGRPC = semconv.RPCSystemKey.String("grpc")
	// RPCNameMessage 是名为 message 的消息的语义约定。
	RPCNameMessage = RPCNameKey.String("message")
	// RPCMessageTypeSent 是已发送的 RPC 消息类型的语义约定。
	RPCMessageTypeSent = RPCMessageTypeKey.String("SENT")
	// RPCMessageTypeReceived 是已接收的 RPC 消息类型的语义约定。
	RPCMessageTypeReceived = RPCMessageTypeKey.String("RECEIVED")
)

// StatusCodeAttr 返回表示给定代码 c 的 attribute.KeyValue。
func StatusCodeAttr(c gcodes.Code) attribute.KeyValue {
	return GRPCStatusCodeKey.Int64(int64(c))
}
