package trace

// TraceName 表示一次跟踪的名称。
const TraceName = "god"

// Config 表示一个 opentelemetry 配置项。
type Config struct {
	Name     string  `json:",optional"`
	Endpoint string  `json:",optional"`
	Sampler  float64 `json:",default=1.0"`
	Exporter string  `json:",default=jaeger,options=jaeger|zipkin"`
}
