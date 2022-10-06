package trace

// Name 表示跟踪的名称。
const Name = "god"

// Config 是一个 opentelemetry 分布式跟踪的配置。
type Config struct {
	Name     string  `json:",optional"`
	Endpoint string  `json:",optional"`
	Sampler  float64 `json:",default=1.0"`
	Batcher  string  `json:",default=jaeger,options=jaeger|zipkin|grpc"`
}
