package threading

import "sync"

// RoutineGroup 用于将 goroutine 分组在一起，并等待所有 goroutine 完成。
type RoutineGroup struct {
	waitGroup sync.WaitGroup
}

// NewRoutineGroup 返回一个 goroutine 组。
func NewRoutineGroup() *RoutineGroup {
	return new(RoutineGroup)
}

// Run 在 RoutineGroup 内运行指定函数。
// 不要引用外部变量，因为外部变量可能被其他 goroutine 改写。
func (g *RoutineGroup) Run(fn func()) {
	g.waitGroup.Add(1)

	go func() {
		defer g.waitGroup.Done()
		fn()
	}()
}

// RunSafe 在 RoutineGroup 内运行指定函数，并避免 panic。
// 不要引用外部变量，因为外部变量可能被其他 goroutine 改写。
func (g *RoutineGroup) RunSafe(fn func()) {
	g.waitGroup.Add(1)

	GoSafe(func() {
		defer g.waitGroup.Done()
		fn()
	})
}

// Wait 等待所有函数运行完毕。
func (g *RoutineGroup) Wait() {
	g.waitGroup.Wait()
}
