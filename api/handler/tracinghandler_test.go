package handler

import (
	"context"
	"github.com/gotid/god/api/chain"
	gtrace "github.com/gotid/god/lib/trace"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOtelHandler(t *testing.T) {
	gtrace.StartAgent(gtrace.Config{
		Name:     "god-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer gtrace.StopAgent()

	for _, test := range []string{"", "bar"} {
		t.Run(test, func(t *testing.T) {
			h := chain.New(TracingHandler("foo", test)).Then(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					span := trace.SpanFromContext(r.Context())
					assert.True(t, span.SpanContext().IsValid())
					assert.True(t, span.IsRecording())
				}))
			ts := httptest.NewServer(h)
			defer ts.Close()

			client := ts.Client()
			err := func(ctx context.Context) error {
				ctx, span := otel.Tracer("httptrace/client").Start(ctx, "test")
				defer span.End()

				req, _ := http.NewRequest("GET", ts.URL, nil)
				otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

				res, err := client.Do(req)
				assert.Nil(t, err)
				return res.Body.Close()
			}(context.Background())

			assert.Nil(t, err)
		})
	}
}

func TestDontTracingSpanName(t *testing.T) {
	gtrace.StartAgent(gtrace.Config{
		Name:     "god-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})

	DontTraceSpan("bar")

	for _, test := range []string{"", "bar", "foo"} {
		t.Run(test, func(t *testing.T) {
			h := chain.New(TracingHandler("foo", test)).Then(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					span := trace.SpanFromContext(r.Context())
					spanCtx := span.SpanContext()
					if test == "bar" {
						assert.False(t, spanCtx.IsValid())
						assert.False(t, span.IsRecording())
						return
					}

					assert.True(t, span.IsRecording())
					assert.True(t, spanCtx.IsValid())
				}))
			ts := httptest.NewServer(h)
			defer ts.Close()

			client := ts.Client()
			err := func(ctx context.Context) error {
				ctx, span := otel.Tracer("httptrace/client").Start(ctx, "test")
				defer span.End()

				req, _ := http.NewRequest("GET", ts.URL, nil)
				otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

				res, err := client.Do(req)
				assert.Nil(t, err)
				return res.Body.Close()
			}(context.Background())

			assert.Nil(t, err)
		})
	}
}
