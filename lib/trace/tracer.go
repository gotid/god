package trace

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

var _ propagation.TextMapCarrier = new(metadataSupplier)

type metadataSupplier struct {
	metadata *metadata.MD
}

func (s *metadataSupplier) Get(key string) string {
	vals := s.metadata.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
}

func (s *metadataSupplier) Set(key string, val string) {
	s.metadata.Set(key, val)
}

func (s *metadataSupplier) Keys() []string {
	out := make([]string, 0, len(*s.metadata))
	for key := range *s.metadata {
		out = append(out, key)
	}
	return out
}

// Inject 注入元数据至上下文。
func Inject(ctx context.Context, p propagation.TextMapPropagator, md *metadata.MD) {
	p.Inject(ctx, &metadataSupplier{metadata: md})
}

// Extract 从上下文提取元数据。
func Extract(ctx context.Context, p propagation.TextMapPropagator, md *metadata.MD) (baggage.Baggage,
	oteltrace.SpanContext) {
	ctx = p.Extract(ctx, &metadataSupplier{metadata: md})

	return baggage.FromContext(ctx), oteltrace.SpanContextFromContext(ctx)
}
