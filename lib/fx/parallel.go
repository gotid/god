package fx

import "git.zc0901.com/go/god/lib/threading"

// Parallel 并行运行一组函数并等待完成。
func Parallel(fns ...func()) {
	group := threading.NewRoutineGroup()
	for _, fn := range fns {
		group.RunSafe(fn)
	}
	group.Wait()
}
