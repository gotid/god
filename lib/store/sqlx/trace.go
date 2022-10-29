package sqlx

import (
	"context"
	"database/sql"

	"github.com/gotid/god/lib/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// 用于标识 SQL 执行的跨度名称。
const spanName = "sql"

var sqlAttributeKey = attribute.Key("sql.method")

func startSpan(ctx context.Context, method string) (context.Context, oteltrace.Span) {
	tracer := otel.GetTracerProvider().Tracer(trace.Name)
	start, span := tracer.Start(ctx, spanName, oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	span.SetAttributes(sqlAttributeKey.String(method))

	return start, span
}

func endSpan(span oteltrace.Span, err error) {
	defer span.End()

	if err == nil || err == sql.ErrNoRows {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}
