package mathx

// MinInt 返回 a, b 中较小的一个。
func MinInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// MaxInt 返回 a, b 中较大的一个。
func MaxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
