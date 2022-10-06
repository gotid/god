package threading

import "sync"

// RoutineGroup 用于等待一组协程全部执行完毕。
type RoutineGroup struct {
	waitGroup sync.WaitGroup
}

// NewRoutineGroup 返回一个 RoutineGroup。
func NewRoutineGroup() *RoutineGroup {
	return new(RoutineGroup)
}

// Run 运行 RoutineGroup 中的给定函数 fn。
// 不要从外部引用变量，因为可能被其他协程更改。
func (g *RoutineGroup) Run(fn func()) {
	g.waitGroup.Add(1)

	go func() {
		defer g.waitGroup.Done()
		fn()
	}()
}

// RunSafe 运行给定函数 fn，若 panics 则记录。
// 不要从外部引用变量，因为可能被其他协程更改。
func (g *RoutineGroup) RunSafe(fn func()) {
	g.waitGroup.Add(1)

	GoSafe(func() {
		defer g.waitGroup.Done()
		fn()
	})
}

// Wait 等待所有运行中的函数执行完毕。
func (g *RoutineGroup) Wait() {
	g.waitGroup.Wait()
}
