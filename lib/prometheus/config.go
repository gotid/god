package prometheus

// Config 是一个普罗米修斯 prometheus 配置。
type Config struct {
	Host string `json:",optional"`
	Port int    `json:",default=9001"`
	Path string `json:",default=/metrics"`
}
