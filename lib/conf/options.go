package conf

type (
	// Option 是自定义配置项的方法。
	Option func(opt *options)

	options struct {
		env bool
	}
)

// UseEnv 使用环境变量。
func UseEnv() Option {
	return func(opt *options) {
		opt.env = true
	}
}
