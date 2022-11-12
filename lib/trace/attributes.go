package trace

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	gcodes "google.golang.org/grpc/codes"
)

const (
	// GrpcStatusCodeKey 是 Grpc 请求的数字状态码的约定方式。
	GrpcStatusCodeKey = attribute.Key("rpc.grpc.status_code")
	// RpcNameKey 是传输或接收的消息名称。
	RpcNameKey = attribute.Key("name")
	// RpcMessageTypeKey 是传输或接收的消息类型。
	RpcMessageTypeKey = attribute.Key("message.type")
	// RpcMessageIDKey 是传输或接收的消息标识符。
	RpcMessageIDKey = attribute.Key("message.id")
	// RPCMessageCompressedSizeKey 是传输或接收的消息压缩后的字节大小。
	RPCMessageCompressedSizeKey = attribute.Key("message.compressed_size")
	// RpcMessageUncompressedSizeKey 是传输或接收的消息未压缩的字节大小。
	RpcMessageUncompressedSizeKey = attribute.Key("message.uncompressed_size")
)

// 常见 RPC 属性的语义约定。
var (
	// RpcSystemGrpc 是 Grpc 作为远程处理系统的语义约定。
	RpcSystemGrpc = semconv.RPCSystemKey.String("grpc")
	// RpcNameMessage 是名为 message 的消息的语义约定。
	RpcNameMessage = RpcNameKey.String("message")
	// RPCMessageTypeSent 是已发送的 RPC 消息类型的语义约定。
	RPCMessageTypeSent = RpcMessageTypeKey.String("SENT")
	// RPCMessageTypeReceived 是已接收的 RPC 消息类型的语义约定。
	RPCMessageTypeReceived = RpcMessageTypeKey.String("RECEIVED")
)

// StatusCodeAttr 返回表示给定代码 c 的 attribute.KeyValue。
func StatusCodeAttr(c gcodes.Code) attribute.KeyValue {
	return GrpcStatusCodeKey.Int64(int64(c))
}
