package neo

import (
	"errors"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type (
	// Scanner 是一个 neo4j.Result 的结果回调函数。
	Scanner func(result neo4j.Result) error
)

var (
	ErrNotSettable          = errors.New("扫描目标不可设置")
	ErrNotReadableValue     = errors.New("neo: 无法读取的值，检查结构字段是否大写开头")
	ErrUnsupportedValueType = errors.New("neo: 不支持的扫描目标类型")
)

// Label 定义标签类型
type Label string

// 返回标签类型的字符串形式。
func (l Label) String() string {
	return string(l)
}

// RelationshipType 定义关系类型
type RelationshipType string

// 返回关系类型的字符串形式。
func (r RelationshipType) String() string {
	return string(r)
}

const (
	// All 默认为所有关系。
	All RelationshipType = ""
	// View 浏览关系
	View RelationshipType = "VIEW"
	// Down 下载关系
	Down RelationshipType = "DOWN"
	// Fav 收藏关系
	Fav RelationshipType = "FAV"
)

// Direction 定义关系方向类型
type Direction string

const (
	Both     Direction = "-"
	Outgoing Direction = "->"
	Incoming Direction = "<-"
)

// ReverseDirection 返回翻转后的方向。
func ReverseDirection(d Direction) Direction {
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
