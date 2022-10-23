package serverinterceptors

import (
	"context"
	gtrace "github.com/gotid/god/lib/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryTracingInterceptor 用于一元请求的 opentelemetry 链路跟踪拦截器。
func UnaryTracingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := startSpan(ctx, info.FullMethod)
	defer span.End()

	gtrace.MessageReceived.Event(ctx, 1, req)
	resp, err := handler(ctx, req)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(gtrace.StatusCodeAttr(s.Code()))
			gtrace.MessageSent.Event(ctx, 1, s.Proto())
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	span.SetAttributes(gtrace.StatusCodeAttr(gcodes.OK))
	gtrace.MessageSent.Event(ctx, 1, resp)

	return resp, nil
}

// StreamTracingInterceptor 用于流式请求的 opentelemetry 链路跟踪拦截器。
func StreamTracingInterceptor(svr interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx, span := startSpan(ss.Context(), info.FullMethod)
	defer span.End()

	if err := handler(svr, wrapServerStream(ctx, ss)); err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(gtrace.StatusCodeAttr(s.Code()))
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return err
	}

	span.SetAttributes(gtrace.StatusCodeAttr(gcodes.OK))
	return nil
}

// 包装 grpc.ServerStream 并拦截 RecvMsg 和 SendMsg 方法调用。
type serverStream struct {
	grpc.ServerStream
	ctx               context.Context
	receivedMessageID int
	sentMessageID     int
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}

func (s *serverStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.receivedMessageID++
		gtrace.MessageReceived.Event(s.Context(), s.receivedMessageID, m)
	}

	return err
}

func (s *serverStream) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	s.sentMessageID++
	gtrace.MessageSent.Event(s.Context(), s.sentMessageID, m)

	return err
}

// 使用给定的上下文包装 grpc.ServerStream。
func wrapServerStream(ctx context.Context, ss grpc.ServerStream) *serverStream {
	return &serverStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

func startSpan(ctx context.Context, method string) (context.Context, trace.Span) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	bags, spanCtx := gtrace.Extract(ctx, otel.GetTextMapPropagator(), &md)
	ctx = baggage.ContextWithBaggage(ctx, bags)
	tr := otel.Tracer(gtrace.Name)
	name, attrs := gtrace.SpanInfo(method, gtrace.PeerFromCtx(ctx))

	return tr.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), name,
		trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attrs...))
}
