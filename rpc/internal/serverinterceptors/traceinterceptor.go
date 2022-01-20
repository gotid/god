package serverinterceptors

import (
	"context"

	gcodes "google.golang.org/grpc/codes"

	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"

	godtrace "github.com/gotid/god/lib/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryTraceInterceptor 是一个基于 opentelemetry 的 grpc.UnaryServerInterceptor。
func UnaryTraceInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := startSpan(ctx, info.FullMethod)
	defer span.End()

	godtrace.MessageReceived.Event(ctx, 1, req)
	resp, err := handler(ctx, req)
	if err != nil {
		s, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(godtrace.StatusCodeAttr(s.Code()))
			godtrace.MessageSent.Event(ctx, 1, s.Proto())
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		return nil, err
	}

	span.SetAttributes(godtrace.StatusCodeAttr(gcodes.OK))
	godtrace.MessageSent.Event(ctx, 1, resp)

	return resp, nil
}

// StreamTracingInterceptor 返回一个基于 opentelemetry 的 grpc.StreamServerInterceptor。
func StreamTracingInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	ctx, span := startSpan(ss.Context(), info.FullMethod)
	defer span.End()

	if err := handler(srv, wrapServerStream(ctx, ss)); err != nil {
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

// serverStream wraps around the embedded grpc.ServerStream,
// and intercepts the RecvMsg and SendMsg method call.
type serverStream struct {
	grpc.ServerStream
	ctx               context.Context
	receivedMessageID int
	sentMessageID     int
}

func (w *serverStream) Context() context.Context {
	return w.ctx
}

func (w *serverStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err == nil {
		w.receivedMessageID++
		godtrace.MessageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}

func (w *serverStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)
	w.sentMessageID++
	godtrace.MessageSent.Event(w.Context(), w.sentMessageID, m)

	return err
}

// wrapServerStream wraps the given grpc.ServerStream with the given context.
func wrapServerStream(ctx context.Context, ss grpc.ServerStream) *serverStream {
	return &serverStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

func startSpan(ctx context.Context, method string) (context.Context, trace.Span) {
	var md metadata.MD
	reqMetadata, ok := metadata.FromIncomingContext(ctx)
	if ok {
		md = reqMetadata.Copy()
	} else {
		md = metadata.MD{}
	}
	bags, spanCtx := godtrace.Extract(ctx, otel.GetTextMapPropagator(), &md)
	ctx = baggage.ContextWithBaggage(ctx, bags)
	tr := otel.Tracer(godtrace.TraceName)
	name, attr := godtrace.SpanInfo(method, godtrace.PeerFromCtx(ctx))

	return tr.Start(trace.ContextWithRemoteSpanContext(ctx, spanCtx), name,
		trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attr...))
}
