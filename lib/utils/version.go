package utils

import (
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/stringx"
	"strconv"
	"strings"
)

var replacer = stringx.NewReplacer(map[string]string{
	"V": "",
	"v": "",
	"-": ".",
})

// CompareVersions 如果符合 "v1 op v2" 表达式，则返回真。
func CompareVersions(v1, op, v2 string) bool {
	result := compare(v1, v2)
	switch op {
	case "=", "==":
		return result == 0
	case "<":
		return result == -1
	case ">":
		return result == 1
	case "<=":
		return result == -1 || result == 0
	case ">=":
		return result == 1 || result == 0
	}

	return false
}

// v1 < v2 返回 -1，v1 > v2 返回 1，其他返回 0
func compare(v1, v2 string) int {
	v1, v2 = replacer.Replace(v1), replacer.Replace(v2)
	fields1, fields2 := strings.Split(v1, "."), strings.Split(v2, ".")
	ver1, ver2 := strToInts(fields1), strToInts(fields2)
	ver1len, ver2len := len(ver1), len(ver2)
	shorter := mathx.MinInt(ver1len, ver2len)

	for i := 0; i < shorter; i++ {
		if ver1[i] == ver2[i] {
			continue
		} else if ver1[i] < ver2[i] {
			return -1
		} else {
			return 1
		}
	}

	if ver1len < ver2len {
		return -1
	} else if ver1len > ver2len {
		return 1
	} else {
		return 0
	}
}

func strToInts(strs []string) []int64 {
	if len(strs) == 0 {
		return nil
	}

	ret := make([]int64, 0, len(strs))
	for _, str := range strs {
		i, _ := strconv.ParseInt(str, 10, 64)
		ret = append(ret, i)
	}

	return ret
}
