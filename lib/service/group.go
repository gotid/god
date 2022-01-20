package service

import (
	"log"

	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
)

type (
	// Starter 是一个定义 Start 开始方法的接口。
	Starter interface {
		Start()
	}

	// Stopper 是一个定义 Stop 停止方法的接口。
	Stopper interface {
		Stop()
	}

	// Service 是一个组合定义开始和停止方法的接口。
	Service interface {
		Starter
		Stopper
	}

	// Group 是一个服务组。
	// 注意：不保证添加服务端启动顺序。
	Group struct {
		services []Service
		stopOnce func()
	}
)

// NewServiceGroup 返回一个服务组。
func NewServiceGroup() *Group {
	g := new(Group)
	g.stopOnce = syncx.Once(g.doStop)
	return g
}

// Add 添加一个服务到服务组。
func (g *Group) Add(service Service) {
	// 将新增服务添加到最前，按反向顺序停止。
	g.services = append([]Service{service}, g.services...)
}

// Start 启用该服务组。
// 调用该方法后不应有任何逻辑代码，因为该方法是阻塞的，
// 同时，退出该方法后将关闭 logx 输出。
func (g *Group) Start() {
	proc.AddShutdownListener(func() {
		log.Println("服务关闭中...")
		g.stopOnce()
	})

	g.doStart()
}

// Stop 关闭该服务组。
func (g *Group) Stop() {
	g.stopOnce()
}

func (g *Group) doStart() {
	routineGroup := threading.NewRoutineGroup()

	for i := range g.services {
		service := g.services[i]
		routineGroup.RunSafe(func() {
			service.Start()
		})
	}

	routineGroup.Wait()
}

func (g *Group) doStop() {
	for _, service := range g.services {
		service.Stop()
	}
}

// WithStart 将指定函数包装为一个服务。
func WithStart(start func()) Service {
	return startOnlyService{
		start: start,
	}
}

// WithStarter 将指定 Starter 包装为一个服务。
func WithStarter(start Starter) Service {
	return starterOnlyService{
		Starter: start,
	}
}

type (
	stopper struct{}

	startOnlyService struct {
		start func()
		stopper
	}

	starterOnlyService struct {
		Starter
		stopper
	}
)

func (s stopper) Stop() {
}

func (s startOnlyService) Start() {
	s.start()
}
