package trace

import (
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"git.zc0901.com/go/god/lib/logx"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	kindJaeger = "jaeger"
	kindZipkin = "zipkin"
)

var once sync.Once

// StartAgent 启动一个 opentelemetry 代理。
func StartAgent(c Config) {
	once.Do(func() {
		startAgent(c)
	})
}

func startAgent(c Config) {
	opts := []sdktrace.TracerProviderOption{
		// 将基于父范围的采样率设置为100%
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Sampler))),
		// 记录该应用有关信息。
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(c.Name))),
	}

	if len(c.Endpoint) > 0 {
		exporter, err := createExporter(c)
		if err != nil {
			logx.Error(err)
			return
		}

		// 务必在生产环境中开启批量。
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logx.Errorf("[otel] 错误：%v", err)
	}))
}

func createExporter(c Config) (sdktrace.SpanExporter, error) {
	// 当前仅支持 jaeger
	switch c.Exporter {
	case kindJaeger:
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.Endpoint)))
	case kindZipkin:
		return zipkin.New(c.Endpoint)
	default:
		return nil, fmt.Errorf("未知 Exporter: %s", c.Exporter)
	}
}
