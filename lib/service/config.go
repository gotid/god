package service

import (
	"git.zc0901.com/go/god/lib/load"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/prometheus"
	"git.zc0901.com/go/god/lib/stat"
	"github.com/prometheus/common/log"
)

const (
	DevMode  = "dev"  // 开发模式
	TestMode = "test" // 测试环境
	PreMode  = "pre"  // 预发布模式
	ProMode  = "pro"  // 生产模式
)

type Conf struct {
	Name       string            // 服务名称
	Log        logx.LogConf      // 日志配置
	Mode       string            `json:",default=pro,options=dev|test|pre|pro"` // 服务环境，dev-开发环境，test-测试环境，pre-预发环境，pro-正式环境
	MetricsUrl string            `json:",optional"`                             // 指标上报接口地址，该地址需要支持 post json 即可
	Prometheus prometheus.Config `json:",optional"`                             // 普罗米修斯配置
}

func (c Conf) MustSetup() {
	if err := c.Setup(); err != nil {
		log.Fatal(err)
	}
}

// Setup 设置并初始化服务配置（初始化启动模式、普罗米修斯代理、统计输出器等）
func (c Conf) Setup() error {
	if len(c.Log.ServiceName) == 0 {
		c.Log.ServiceName = c.Name
	}

	// 初始化日志
	if err := logx.Setup(c.Log); err != nil {
		return err
	}

	// 非生产模式禁用负载均衡和日志汇报
	c.initMode()

	// 方式一：启动普罗米修斯http服务端口
	prometheus.StartAgent(c.Prometheus)

	// 方式二：设置统计报告书写器（写入普罗米修斯）
	if len(c.MetricsUrl) > 0 {
		stat.SetReportWriter(stat.NewRemoteWriter(c.MetricsUrl))
	}

	return nil
}

func (c Conf) initMode() {
	switch c.Mode {
	case DevMode, TestMode, PreMode:
		// 开发、测试、预发布模式，不启用负载均衡和统计上报
		load.Disable()
		stat.SetReporter(nil)
	}
}
