package devserver

// Config 用于内部 http 服务器的配置。
type Config struct {
	Enable        bool   `json:",default=true"`
	Host          string `json:",optional"`
	Port          int    `json:",default=6470"`
	MetricsPath   string `json:",default=/metrics"`
	HealthPath    string `json:",default=/healthz"`
	EnableMetrics bool   `json:",default=true"`
	EnablePprof   bool   `json:",default=true"`
}
