package service

import (
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
	"log"
)

type (
	// Starter 是包装 Start 方法的接口。
	Starter interface {
		Start()
	}

	// Stopper 是包装 Stop 方法的接口。
	Stopper interface {
		Stop()
	}

	// Service 是既有 Start 又有 Stop 方法的接口。
	Service interface {
		Starter
		Stopper
	}

	// Group 是一组 Service。
	// 注意：不保证添加服务的启动顺序。
	Group struct {
		services []Service
		stopOnce func()
	}
)

// NewGroup 返回一个 Group。
func NewGroup() *Group {
	g := new(Group)
	g.stopOnce = syncx.Once(g.doStop)
	return g
}

// Add 添加服务到群组。
func (g *Group) Add(service Service) {
	// 加到最前，按相反顺序停止
	g.services = append([]Service{service}, g.services...)
}

// Start 启动一组服务。
// 注意：在调用该方法后不应有任何逻辑代码，因为该方法是阻塞的。
// 另外，退出该方法时将关闭 logx 输出。
func (g *Group) Start() {
	proc.AddShutdownListener(func() {
		log.Println("一组服务关闭中...")
		g.stopOnce()
	})

	g.doStart()
}

// Stop 关闭一组服务。
func (g *Group) Stop() {
	g.stopOnce()
}

// 使用协程启动一组服务
func (g *Group) doStart() {
	routineGroup := threading.NewRoutineGroup()

	for i := range g.services {
		// 此处用索引不会导致协程错误
		// 或者使用 service := service 重新赋值给一个新变量
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

// WithStart 将启动函数 start 包装为一个 Service 服务。
func WithStart(start func()) Service {
	return startOnlyService{
		start: start,
	}
}

// WithStarter 将 Starter 接口的实现包装为一个 Service 服务。
func WithStarter(starter Starter) Service {
	return starterOnlyService{
		Starter: starter,
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
