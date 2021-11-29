package neo

import (
	"errors"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type (
	// Scanner 是一个 neo4j.Result 的结果回调函数。
	Scanner func(result neo4j.Result) error
)

var (
	ErrNotSettable          = errors.New("[Neo] 扫描目标不可设置")
	ErrNotReadableValue     = errors.New("[Neo] 无法读取的值，检查结构字段是否大写开头")
	ErrUnsupportedValueType = errors.New("[Neo] 不支持的扫描目标类型")
)

// Label 定义标签类型
type Label string

// 返回标签类型的字符串形式。
func (l Label) String() string {
	return string(l)
}

// Labels 定义一个标签切片类型
type Labels []Label

// FromLabels 返回一个标签切片组。
func FromLabels(ls ...string) Labels {
	ret := Labels{}
	for i := range ls {
		ret = append(ret, Label(ls[i]))
	}
	return ret
}

// Stringify 返回字符串化的标签切片形式。
func (labels Labels) Stringify() []string {
	ret := make([]string, 0)
	for _, label := range labels {
		ret = append(ret, label.String())
	}
	return ret
}

// String 返回字符串化的标签特征形式。
func (labels Labels) String() string {
	ls := labels.Stringify()
	return strings.Join(ls, ":")
}

// RelationType 定义关系类型
type RelationType string

// 返回关系类型的字符串形式。
func (r RelationType) String() string {
	return string(r)
}

const (
	// All 默认为所有关系。
	All RelationType = ""
	// View 浏览关系
	View RelationType = "VIEW"
	// Down 下载关系
	Down RelationType = "DOWN"
	// Fav 收藏关系
	Fav RelationType = "FAV"
)

// Direction 定义关系方向类型
type Direction string

const (
	Both     Direction = "-"
	Outgoing Direction = "->"
	Incoming Direction = "<-"
)

// Reverse 返回翻转后的方向。
func (d Direction) Reverse() Direction {
	switch d {
	case Both:
		return Both
	case Outgoing:
		return Incoming
	case Incoming:
		return Outgoing
	default:
		panic("不支持的 Neo4j Direction")
	}
}
