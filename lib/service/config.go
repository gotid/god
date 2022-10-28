package service

import (
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/prometheus"
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/lib/trace"
	"log"
)

const (
	DevMode  = "dev"  // 开发模式
	TestMode = "test" // 测试模式
	RtMode   = "rt"   // 回归测试模式
	PreMode  = "pre"  // 预发布模式
	ProMode  = "pro"  // 生产模式
)

// Config 是一个服务配置。
type Config struct {
	Name       string
	Log        logx.Config
	Mode       string            `json:",default=pro,options=[dev,tet,rt,pre,pro]"`
	MetricsUrl string            `json:",optional"`
	Prometheus prometheus.Config `json:",optional"`
	Telemetry  trace.Config      `json:",optional"`
}

// MustSetup 设置服务，出错退出。
func (c Config) MustSetup() {
	if err := c.Setup(); err != nil {
		log.Fatal(err)
	}
}

// Setup 设置服务。
func (c Config) Setup() error {
	if len(c.Log.ServiceName) == 0 {
		c.Log.ServiceName = c.Name
	}
	if err := logx.Setup(c.Log); err != nil {
		return err
	}

	c.initMode()

	prometheus.StartAgent(c.Prometheus)

	if len(c.Telemetry.Name) == 0 {
		c.Telemetry.Name = c.Name
	}
	trace.StartAgent(c.Telemetry)
	proc.AddShutdownListener(func() {
		trace.StopAgent()
	})

	if len(c.MetricsUrl) > 0 {
		stat.SetReportWriter(stat.NewRemoteWriter(c.MetricsUrl))
	}

	return nil
}

func (c Config) initMode() {
	switch c.Mode {
	case DevMode, TestMode, RtMode, PreMode:
		load.Disable()
		stat.SetReporter(nil)
	}
}
