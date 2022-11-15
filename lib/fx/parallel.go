package fx

import "github.com/gotid/god/lib/threading"

// Parallel 并行运行 fns 并等待完成。
func Parallel(fns ...func()) {
	group := threading.NewRoutineGroup()
	for _, fn := range fns {
		group.RunSafe(fn)
	}
	group.Wait()
}
