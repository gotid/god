package queue

import "strings"

// 生成 MultiPusher 的名称。
func generateName(pushers []Pusher) string {
	names := make([]string, len(pushers))
	for i, pusher := range pushers {
		names[i] = pusher.Name()
	}

	return strings.Join(names, ",")
}