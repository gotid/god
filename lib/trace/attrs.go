package trace

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	gcodes "google.golang.org/grpc/codes"
)

const (
	// GRPCStatusCodeKey gRPC 响应状态码的约定方法。
	GRPCStatusCodeKey = attribute.Key("rpc.grpc.status_code")
	// RPCNameKey 发送或接收的消息名称。
	RPCNameKey = attribute.Key("name")
	// RPCMessageTypeKey 发送或接收的消息类型。
	RPCMessageTypeKey = attribute.Key("message.type")
	// RPCMessageIDKey 发送或接收的消息ID。
	RPCMessageIDKey = attribute.Key("message.id")
	// RPCMessageCompressedSizeKey 发送或接收的消息压缩后大小。
	RPCMessageCompressedSizeKey = attribute.Key("message.compressed_size")
	// RPCMessageUncompressedSizeKey 发送或接收的消息未压缩大小。
	RPCMessageUncompressedSizeKey = attribute.Key("message.uncompressed_size")
)

var (
	// RPCSystemGRPC gRPC 作为远程系统的语义约定。
	RPCSystemGRPC = semconv.RPCSystemKey.String("grpc")
	// RPCNameMessage 名为 message 的消息在语义上的约定。
	RPCNameMessage = RPCNameKey.String("message")
	// RPCMessageTypeSent 发送 RPC 消息的类型语义约定。
	RPCMessageTypeSent = RPCMessageTypeKey.String("SENT")
	// RPCMessageTypeReceived 接收 RPC 消息的类型语义约定。
	RPCMessageTypeReceived = RPCMessageTypeKey.String("RECEIVED")
)

// StatusCodeAttr 返回表示指定 grpc 代码的 attribute.KeyValue。
func StatusCodeAttr(code gcodes.Code) attribute.KeyValue {
	return GRPCStatusCodeKey.Int64(int64(code))
}
