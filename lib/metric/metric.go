package metric

// VectorOpts 是一个向量选项的通用配置。
type VectorOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}
