package clientinterceptors

import (
	"context"

	gcodes "google.golang.org/grpc/codes"

	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"

	godtrace "git.zc0901.com/go/god/lib/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryTraceInterceptor 返回一个 opentelemetry 的 grpc.UnaryClientInterceptor 拦截器。
func UnaryTraceInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// 开启客户端跟踪操作
	ctx, span := startSpan(ctx, method, cc.Target())
	defer span.End()

	godtrace.MessageSent.Event(ctx, 1, req)
	godtrace.MessageReceived.Event(ctx, 1, reply)

	if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(godtrace.StatusCodeAttr(s.Code()))
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	span.SetAttributes(godtrace.StatusCodeAttr(gcodes.OK))

	return nil
}

func startSpan(ctx context.Context, method, target string) (context.Context, trace.Span) {
	var md metadata.MD
	reqMetadata, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = reqMetadata.Copy()
	} else {
		md = metadata.MD{}
	}

	tr := otel.Tracer(godtrace.TraceName)
	name, attr := godtrace.SpanInfo(method, target)
	ctx, span := tr.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attr...))
	godtrace.Inject(ctx, otel.GetTextMapPropagator(), &md)
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx, span
}
