package trace

import (
	"context"
	"fmt"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"sync"
)

const (
	kindJaeger = "jaeger"
	kindZipkin = "zipkin"
	kindGrpc   = "grpc"
)

var (
	agents = make(map[string]lang.PlaceholderType)
	lock   sync.Mutex
	tp     *sdktrace.TracerProvider
)

// StartAgent 启动一个 opentelemetry 跟踪代理。
func StartAgent(c Config) {
	lock.Lock()
	defer lock.Unlock()

	_, ok := agents[c.Endpoint]
	if ok {
		return
	}

	// 如果出错，让之后的调用运行。
	if err := startAgent(c); err != nil {
		return
	}

	agents[c.Endpoint] = lang.Placeholder
}

// StopAgent 按注册顺序关闭 span 处理器。
func StopAgent() {
	_ = tp.Shutdown(context.Background())
}
func startAgent(c Config) error {
	opts := []sdktrace.TracerProviderOption{
		// 基于父跨度span的采样率设置为100%
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Sampler))),
		// 在资源中记录有关此应用程序的信息
		sdktrace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(c.Name))),
	}

	if len(c.Endpoint) > 0 {
		exporter, err := createExporter(c)
		if err != nil {
			logx.Error(err)
			return err
		}

		// 务必在生产环境中进行批量操作
		opts = append(opts, sdktrace.WithBatcher(exporter))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logx.Errorf("[otel] 错误：%v", err)
	}))

	return nil
}

func createExporter(c Config) (sdktrace.SpanExporter, error) {
	// 现在只支持 jaeger 和 zipkin，以后会支持更多
	switch c.Batcher {
	case kindJaeger:
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.Endpoint)))
	case kindZipkin:
		return zipkin.New(c.Endpoint)
	case kindGrpc:
		return otlptracegrpc.NewUnstarted(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(c.Endpoint),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		), nil
	default:
		return nil, fmt.Errorf("未知的 exportor: %s", c.Batcher)
	}
}
