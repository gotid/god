package trace

import (
	"context"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

type metadataSupplier struct {
	metadata *metadata.MD
}

func (s *metadataSupplier) Get(key string) string {
	values := s.metadata.Get(key)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func (s *metadataSupplier) Set(key, value string) {
	s.metadata.Set(key, value)
}

func (s *metadataSupplier) Keys() []string {
	out := make([]string, 0, len(*s.metadata))
	for key := range *s.metadata {
		out = append(out, key)
	}

	return out
}

// Inject 将 ctx 中的横切关注点注入上下文。
func Inject(ctx context.Context, p propagation.TextMapPropagator, md *metadata.MD) {
	p.Inject(ctx, &metadataSupplier{
		metadata: md,
	})
}

// Extract 从上下文中提取 metadata。
func Extract(ctx context.Context, p propagation.TextMapPropagator, md *metadata.MD) (baggage.Baggage, oteltrace.SpanContext) {
	ctx = p.Extract(ctx, &metadataSupplier{
		metadata: md,
	})

	return baggage.FromContext(ctx), oteltrace.SpanContextFromContext(ctx)
}
