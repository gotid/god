package util

// TrimStringSlice 返回一个没有空白字符项的切片副本。
func TrimStringSlice(list []string) []string {
	var out []string
	for _, item := range list {
		if len(item) == 0 {
			continue
		}
		out = append(out, item)
	}

	return out
}
