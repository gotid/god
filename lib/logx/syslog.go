package logx

import "log"

type redirector struct{}

// CollectSysLog 将系统日志重定向到 logx info
func CollectSysLog() {
	log.SetOutput(new(redirector))
}

func (r *redirector) Write(p []byte) (n int, err error) {
	Info(string(p))
	return len(p), nil
}
