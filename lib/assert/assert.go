package assert

import (
	"git.zc0901.com/go/god/internal/empty"

	"git.zc0901.com/go/god/lib/g"
)

func IsTrue(exp bool, msg ...string) {
	m := "[断言失败] - 表达式必须为真"
	if len(msg) == 1 {
		m = msg[0]
	}

	if !exp {
		panic(m)
	}
}

func IsFalse(exp bool, msg ...string) {
	m := "[断言失败] - 表达式必须为假"
	if len(msg) == 1 {
		m = msg[0]
	}

	if exp {
		panic(m)
	}
}

func IsNil(o interface{}, msg ...string) {
	m := "[断言失败] - 对象必须为空"
	if len(msg) == 1 {
		m = msg[0]
	}

	if !g.IsNil(o) {
		panic(m)
	}
}

func IsNotNil(o interface{}, msg ...string) {
	m := "[断言失败] - 对象必须非空"
	if len(msg) == 1 {
		m = msg[0]
	}

	if g.IsNil(o) {
		panic(m)
	}
}

func IsNotEmpty(o interface{}, msg ...string) {
	m := "[断言失败] - 对象必须为非空值"
	if len(msg) == 1 {
		m = msg[0]
	}

	if empty.IsEmpty(o) {
		panic(m)
	}
}

func HasLength(v string, msg ...string) {
	m := "[断言失败] - 字符串必须有长度"
	if len(msg) == 1 {
		m = msg[0]
	}

	if len(v) == 0 {
		panic(m)
	}
}

func IsAll(vs []interface{}, msg ...string) {
	m := "[断言失败] - 必须为非空数组"
	if len(msg) == 1 {
		m = msg[0]
	}

	for _, v := range vs {
		if g.IsNil(v) {
			panic(m)
		}
	}
}
