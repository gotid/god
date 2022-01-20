package clientinterceptors

import (
	"context"
	"io"

	gcodes "google.golang.org/grpc/codes"

	"go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"

	godtrace "github.com/gotid/god/lib/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	receiveEndEvent streamEventType = iota
	errorEvent
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

// StreamTraceInterceptor 返回一个 opentelemetry 的 grpc.StreamClientInterceptor 拦截器。
func StreamTraceInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
	method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx, span := startSpan(ctx, method, cc.Target())
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(godtrace.StatusCodeAttr(st.Code()))
		} else {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
		return s, err
	}

	stream := wrapClientStream(ctx, s, desc)

	go func() {
		if err := <-stream.Finished; err != nil {
			s, ok := status.FromError(err)
			if ok {
				span.SetStatus(codes.Error, s.Message())
				span.SetAttributes(godtrace.StatusCodeAttr(s.Code()))
			} else {
				span.SetStatus(codes.Error, err.Error())
			}
		} else {
			span.SetAttributes(godtrace.StatusCodeAttr(gcodes.OK))
		}

		span.End()
	}()

	return stream, nil
}

type (
	streamEventType int

	streamEvent struct {
		Type streamEventType
		Err  error
	}

	clientStream struct {
		grpc.ClientStream
		Finished          chan error
		desc              *grpc.StreamDesc
		events            chan streamEvent
		eventsDone        chan struct{}
		receivedMessageID int
		sentMessageID     int
	}
)

func (w *clientStream) CloseSend() error {
	err := w.ClientStream.CloseSend()
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return err
}

func (w *clientStream) Header() (metadata.MD, error) {
	md, err := w.ClientStream.Header()
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return md, err
}

func (w *clientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)
	if err == nil && !w.desc.ServerStreams {
		w.sendStreamEvent(receiveEndEvent, nil)
	} else if err == io.EOF {
		w.sendStreamEvent(receiveEndEvent, nil)
	} else if err != nil {
		w.sendStreamEvent(errorEvent, err)
	} else {
		w.receivedMessageID++
		godtrace.MessageReceived.Event(w.Context(), w.receivedMessageID, m)
	}

	return err
}

func (w *clientStream) SendMsg(m interface{}) error {
	err := w.ClientStream.SendMsg(m)
	w.sentMessageID++
	godtrace.MessageSent.Event(w.Context(), w.sentMessageID, m)
	if err != nil {
		w.sendStreamEvent(errorEvent, err)
	}

	return err
}

func (w *clientStream) sendStreamEvent(eventType streamEventType, err error) {
	select {
	case <-w.eventsDone:
	case w.events <- streamEvent{Type: eventType, Err: err}:
	}
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

// wrapClientStream wraps s with given ctx and desc.
func wrapClientStream(ctx context.Context, s grpc.ClientStream, desc *grpc.StreamDesc) *clientStream {
	events := make(chan streamEvent)
	eventsDone := make(chan struct{})
	finished := make(chan error)

	go func() {
		defer close(eventsDone)

		for {
			select {
			case event := <-events:
				switch event.Type {
				case receiveEndEvent:
					finished <- nil
					return
				case errorEvent:
					finished <- event.Err
					return
				}
			case <-ctx.Done():
				finished <- ctx.Err()
				return
			}
		}
	}()

	return &clientStream{
		ClientStream: s,
		desc:         desc,
		events:       events,
		eventsDone:   eventsDone,
		Finished:     finished,
	}
}
