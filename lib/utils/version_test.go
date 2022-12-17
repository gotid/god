package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		ver1     string
		operator string
		ver2     string
		except   bool
	}{
		{"1", ">", "1.0.1", false},
		{"1", ">", "0.9.9", true},
		{"1", "<", "1.0-1", true},
		{"1.0.1", "<", "1-0.1", false},
		{"1.0.1", "==", "1.0.1", true},
		{"1.0.1", "==", "1.0.2", false},
		{"1.1-1", "==", "1.0.2", false},
		{"1.0.1", ">=", "1.0.2", false},
		{"1.0.2", ">=", "1.0.2", true},
		{"1.0.3", ">=", "1.0.2", true},
		{"1.0.4", "<=", "1.0.2", false},
		{"1.0.4", "<=", "1.0.6", true},
		{"1.0.4", "<=", "1.0.4", true},
	}

	for _, each := range cases {
		each := each
		t.Run(fmt.Sprintf("%s%s%s", each.ver1, each.operator, each.ver2), func(t *testing.T) {
			actual := CompareVersions(each.ver1, each.operator, each.ver2)
			assert.Equal(t, each.except, actual, fmt.Sprintf("%s vs %s", each.ver1, each.ver2))
		})
	}
}
