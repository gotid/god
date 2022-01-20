package handler

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/gotid/god/lib/trace"
)

// TraceHandler 返回一个处理 opentelemetry 的中间件。
func TraceHandler(serviceName, path string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		propagator := otel.GetTextMapPropagator()
		tracer := otel.GetTracerProvider().Tracer(trace.TraceName)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			spanCtx, span := tracer.Start(
				ctx,
				path,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
				oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(serviceName, path, r)...),
			)
			defer span.End()

			// 设置跟踪编号
			sc := span.SpanContext()
			if sc.HasTraceID() {
				w.Header().Set(trace.TraceIdKey, sc.TraceID().String())
			}

			next.ServeHTTP(w, r.WithContext(spanCtx))
		})
	}
}
