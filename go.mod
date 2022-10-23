module github.com/gotid/god

go 1.16

require (
	github.com/fatih/color v1.13.0
	github.com/stretchr/testify v1.8.0
)

require (
	github.com/alicebob/miniredis/v2 v2.23.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/mock v1.4.4
	github.com/prometheus/client_golang v1.13.0
	github.com/spaolacci/murmur3 v1.1.0
	go.etcd.io/etcd/client/v3 v3.5.5
	go.opentelemetry.io/otel v1.11.0
	go.opentelemetry.io/otel/exporters/jaeger v1.11.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.11.0
	go.opentelemetry.io/otel/exporters/zipkin v1.11.0
	go.opentelemetry.io/otel/sdk v1.11.0
	go.opentelemetry.io/otel/trace v1.11.0
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
	gopkg.in/h2non/gock.v1 v1.1.2
)
