package trace

import (
	"context"
	"go.opentelemetry.io/otel/trace"
)

func SpanIDFromContext(ctx context.Context) string {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasSpanID() {
		return spanContext.SpanID().String()
	}

	return ""
}

func TraceIDFromContext(ctx context.Context) string {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.HasTraceID() {
		return spanContext.TraceID().String()
	}

	return ""
}
