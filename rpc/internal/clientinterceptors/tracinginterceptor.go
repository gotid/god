package clientinterceptors

import (
	"context"
	gtrace "github.com/gotid/god/lib/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	gcodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
)

const (
	receiveEndEvent streamEventType = iota
	errorEvent
)

// UnaryTracingInterceptor 用于一元请求的 opentelemetry 客户端链路跟踪拦截器。
func UnaryTracingInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, span := startSpan(ctx, method, cc.Target())
	defer span.End()

	gtrace.MessageSent.Event(ctx, 1, req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	gtrace.MessageReceived.Event(ctx, 1, reply)
	if err != nil {
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

// StreamTracingInterceptor 用于流式请求的 opentelemetry 客户端链路跟踪拦截器。
func StreamTracingInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
	method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx, span := startSpan(ctx, method, cc.Target())
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(gtrace.StatusCodeAttr(st.Code()))
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
				span.SetAttributes(gtrace.StatusCodeAttr(s.Code()))
			} else {
				span.SetStatus(codes.Error, err.Error())
			}
		} else {
			span.SetAttributes(gtrace.StatusCodeAttr(gcodes.OK))
		}
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

func (s *clientStream) CloseSend() error {
	err := s.ClientStream.CloseSend()
	if err != nil {
		s.sendStreamEvent(errorEvent, err)
	}

	return err
}

func (s *clientStream) Header() (metadata.MD, error) {
	md, err := s.ClientStream.Header()
	if err != nil {
		s.sendStreamEvent(errorEvent, err)
	}

	return md, err
}

func (s *clientStream) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if err == nil && !s.desc.ServerStreams {
		s.sendStreamEvent(receiveEndEvent, nil)
	} else if err == io.EOF {
		s.sendStreamEvent(receiveEndEvent, nil)
	} else if err != nil {
		s.sendStreamEvent(errorEvent, err)
	} else {
		s.receivedMessageID++
		gtrace.MessageReceived.Event(s.Context(), s.receivedMessageID, m)
	}

	return err
}

func (s *clientStream) SendMsg(m interface{}) error {
	err := s.ClientStream.SendMsg(m)
	s.sentMessageID++
	gtrace.MessageSent.Event(s.Context(), s.sentMessageID, m)
	if err != nil {
		s.sendStreamEvent(errorEvent, err)
	}

	return err
}

func (s *clientStream) sendStreamEvent(eventType streamEventType, err error) {
	select {
	case <-s.eventsDone:
	case s.events <- streamEvent{Type: eventType, Err: err}:
	}

}

// 使用上下文和描述包装 grpc.ClientStream。
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
		Finished:     finished,
		desc:         desc,
		events:       events,
		eventsDone:   eventsDone,
	}
}

func startSpan(ctx context.Context, method string, target string) (context.Context, trace.Span) {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	tr := otel.Tracer(gtrace.Name)
	name, attrs := gtrace.SpanInfo(method, target)
	ctx, span := tr.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...))
	gtrace.Inject(ctx, otel.GetTextMapPropagator(), &md)
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx, span
}
