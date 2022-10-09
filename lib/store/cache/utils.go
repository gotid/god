package cache

import "strings"

const keySeparator = ","

// TotalWeights 返回给定节点的总权重。
func TotalWeights(c Config) int {
	var weights int
	for _, node := range c {
		if node.Weight < 0 {
			node.Weight = 0
		}
		weights += node.Weight
	}

	return weights
}

func formatKeys(keys []string) string {
	return strings.Join(keys, keySeparator)
}
