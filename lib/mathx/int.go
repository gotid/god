package mathx

import "golang.org/x/exp/constraints"

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

// Max 返回切片中最大的一个。
func Max[T constraints.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	m := s[0]
	for _, v := range s {
		if m < v {
			m = v
		}
	}

	return m
}

// Min 返回切片中最小的一个。
func Min[T constraints.Ordered](s []T) T {
	if len(s) == 0 {
		var zero T
		return zero
	}

	m := s[0]
	for _, v := range s {
		if m > v {
			m = v
		}
	}

	return m
}
