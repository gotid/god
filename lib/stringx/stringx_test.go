package stringx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"path"
	"testing"
)

func TestJoin(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		expect string
	}{
		{
			name:   "所有元素均为空",
			input:  []string{"", ""},
			expect: "",
		},
		{
			name:   "两个元素",
			input:  []string{"abc", "012"},
			expect: "abc.012",
		},
		{
			name:   "最后一个元素为空",
			input:  []string{"abc", ""},
			expect: "abc",
		},
		{
			name:   "第一个元素为空",
			input:  []string{"", "abc"},
			expect: "abc",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, Join('.', test.input...))
		})
	}
}
func TestRand2(t *testing.T) {
	a, b := Rand(), RandId()
	fmt.Println(a)
	fmt.Println(b)
}
func TestContains(t *testing.T) {
	cases := []struct {
		slice    []string
		value    string
		excepted bool
	}{
		{[]string{"1"}, "1", true},
		{[]string{"1"}, "2", false},
		{[]string{"1", "2"}, "1", true},
		{[]string{"1", "2"}, "3", false},
		{nil, "3", false},
		{nil, "", false},
	}

	for _, each := range cases {
		t.Run(path.Join(each.slice...), func(t *testing.T) {
			actual := Contains(each.slice, each.value)
			assert.Equal(t, each.excepted, actual)
		})
	}
}

func TestFilter(t *testing.T) {
	cases := []struct {
		input    string
		ignores  []rune
		expected string
	}{
		{``, nil, ``},
		{`abc`, nil, `abc`},
		{`ab,cd,ef`, []rune{','}, `abcdef`},
		{`ab, cd,ef`, []rune{',', ' '}, `abcdef`},
		{`ab, cd, ef`, []rune{',', ' '}, `abcdef`},
		{`ab, cd, ef, `, []rune{',', ' '}, `abcdef`},
	}

	for _, each := range cases {
		t.Run(each.input, func(t *testing.T) {
			actual := Filter(each.input, func(r rune) bool {
				for _, x := range each.ignores {
					if x == r {
						return true
					}
				}
				return false
			})
			assert.Equal(t, each.expected, actual)
		})
	}
}

func TestFirstN(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		n        int
		ellipsis string
		expected string
	}{
		{
			name:     "英文字符串",
			input:    "anything that we use",
			n:        8,
			expected: "anything",
		},
		{
			name:     "带省略号的英文字符串",
			input:    "anything that we use",
			n:        8,
			ellipsis: "...",
			expected: "anything...",
		},
		{
			name:     "比英文字符串长度还长的n",
			input:    "anything that we use",
			n:        80,
			expected: "anything that we use",
		},
		{
			name:     "中文字符串",
			input:    "我是中国人",
			n:        2,
			expected: "我是",
		},
		{
			name:     "带省略号的中文字符串",
			input:    "我是中国人",
			n:        2,
			ellipsis: "...",
			expected: "我是...",
		},
		{
			name:     "比中文字符串还长的n",
			input:    "我是中国人",
			n:        20,
			expected: "我是中国人",
		},
	}

	for _, each := range cases {
		t.Run(each.name, func(t *testing.T) {
			assert.Equal(t, each.expected, FirstN(each.input, each.n, each.ellipsis))
		})
	}
}

func TestRemove(t *testing.T) {
	cases := []struct {
		input    []string
		remove   []string
		excepted []string
	}{
		{
			input:    []string{"a", "b", "a", "c"},
			remove:   []string{"a", "b"},
			excepted: []string{"c"},
		},
		{
			input:    []string{"b", "c"},
			remove:   []string{"a"},
			excepted: []string{"b", "c"},
		},
		{
			input:    []string{"b", "a", "c"},
			remove:   []string{"a"},
			excepted: []string{"b", "c"},
		},
		{
			input:    []string{},
			remove:   []string{"a"},
			excepted: []string{},
		},
	}

	for _, each := range cases {
		t.Run(path.Join(each.input...), func(t *testing.T) {
			assert.ElementsMatch(t, each.excepted, Remove(each.input, each.remove...))
		})
	}
}

func TestSubstr(t *testing.T) {
	cases := []struct {
		input    string
		start    int
		stop     int
		err      error
		excepted string
	}{
		{
			input:    "abcdefg",
			start:    1,
			stop:     4,
			excepted: "bcd",
		},
		{
			input:    "我爱北京天安门",
			start:    1,
			stop:     2,
			excepted: "爱",
		},
		{
			input:    "abcdefg",
			start:    -1,
			stop:     4,
			err:      ErrInvalidStartPosition,
			excepted: "",
		},
		{
			input:    "abcdefg",
			start:    100,
			stop:     4,
			err:      ErrInvalidStartPosition,
			excepted: "",
		},
		{
			input:    "abcdefg",
			start:    1,
			stop:     -1,
			err:      ErrInvalidStopPosition,
			excepted: "",
		},
		{
			input:    "abcdefg",
			start:    1,
			stop:     100,
			err:      ErrInvalidStopPosition,
			excepted: "",
		},
	}

	for _, each := range cases {
		t.Run(each.input, func(t *testing.T) {
			val, err := Substr(each.input, each.start, each.stop)
			assert.Equal(t, each.err, err)
			if err == nil {
				assert.Equal(t, each.excepted, val)
			}
		})
	}
}

func TestTakeOne(t *testing.T) {
	cases := []struct {
		valid    string
		or       string
		expected string
	}{
		{"", "", ""},
		{"", "1", "1"},
		{"1", "", "1"},
		{"1", "2", "1"},
	}

	for _, each := range cases {
		t.Run(each.valid, func(t *testing.T) {
			actual := TakeOne(each.valid, each.or)
			assert.Equal(t, each.expected, actual)
		})
	}
}

func TestTakeWithPriority(t *testing.T) {
	tests := []struct {
		fns    []func() string
		expect string
	}{
		{
			fns: []func() string{
				func() string {
					return "first"
				},
				func() string {
					return "second"
				},
				func() string {
					return "third"
				},
			},
			expect: "first",
		},
		{
			fns: []func() string{
				func() string {
					return ""
				},
				func() string {
					return "second"
				},
				func() string {
					return "third"
				},
			},
			expect: "second",
		},
		{
			fns: []func() string{
				func() string {
					return ""
				},
				func() string {
					return ""
				},
				func() string {
					return "third"
				},
			},
			expect: "third",
		},
		{
			fns: []func() string{
				func() string {
					return ""
				},
				func() string {
					return ""
				},
				func() string {
					return ""
				},
			},
			expect: "",
		},
	}

	for _, test := range tests {
		t.Run(RandId(), func(t *testing.T) {
			val := TakeWithPriority(test.fns...)
			assert.Equal(t, test.expect, val)
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			input:  "",
			expect: "",
		},
		{
			input:  "A",
			expect: "a",
		},
		{
			input:  "a",
			expect: "a",
		},
		{
			input:  "hello_world",
			expect: "hello_world",
		},
		{
			input:  "Hello_world",
			expect: "hello_world",
		},
		{
			input:  "hello_World",
			expect: "hello_World",
		},
		{
			input:  "helloWorld",
			expect: "helloWorld",
		},
		{
			input:  "HelloWorld",
			expect: "helloWorld",
		},
		{
			input:  "hello World",
			expect: "hello World",
		},
		{
			input:  "Hello World",
			expect: "hello World",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.expect, ToCamelCase(test.input))
		})
	}
}

func TestUnion(t *testing.T) {
	first := []string{
		"one",
		"two",
		"three",
	}
	second := []string{
		"zero",
		"two",
		"three",
		"four",
	}
	union := Union(first, second)
	contains := func(v string) bool {
		for _, each := range union {
			if v == each {
				return true
			}
		}

		return false
	}
	assert.Equal(t, 5, len(union))
	assert.True(t, contains("zero"))
	assert.True(t, contains("one"))
	assert.True(t, contains("two"))
	assert.True(t, contains("three"))
	assert.True(t, contains("four"))
}
