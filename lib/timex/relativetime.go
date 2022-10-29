package timex

import "time"

// 用足够长的过去时间作为开始时间，以防 timex.Now() - lastTime 等于 0。
var initTime = time.Now().AddDate(-1, -1, -1)

// Now 返回自 initTime 以来的相对时长。
// 调用者只需关心相对值。
func Now() time.Duration {
	return time.Since(initTime)
}

// Since 返回距离 duration 以来的差异。
func Since(t time.Duration) time.Duration {
	return time.Since(initTime) - t
}
