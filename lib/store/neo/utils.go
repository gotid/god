package neo

import (
	"fmt"
	"strings"
	"time"

	"git.zc0901.com/go/god/lib/g"

	"git.zc0901.com/go/god/lib/assert"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// LabelExp 返回 neo4j.Node 标签的特征字符串。
func LabelExp(n neo4j.Node) string {
	return strings.Join(n.Labels, ":")
}

// MustFullNode 必须为完整节点，否则退出程序。
func MustFullNode(node neo4j.Node, name string) {
	assert.IsNotNil(node, name, "节点不能为空")
	assert.IsNotEmpty(node.Id, name, "Id 不能为空")
	assert.IsNotEmpty(LabelExp(node), name, "标签不能为空")
}

// MustFullRelation 必须为完整关系，否则退出程序。
func MustFullRelation(r Relation, name string) {
	assert.IsNotEmpty(r, name, "关系不能为空")
	assert.IsNotEmpty(r.Type, name, "关系类型必须明确")
	assert.IsNotEmpty(r.Direction, name, "关系方向必须明确")
}

// MakeProps 返回 Neo4j 属性字典。
//
// {user: "abc", age: 123}
func MakeProps(params g.Map) string {
	if len(params) == 0 {
		return ""
	}

	b := strings.Builder{}
	b.WriteString("{")
	index := 0
	for k, v := range params {
		index++
		b.WriteString(k)
		b.WriteString(":")
		switch v.(type) {
		case string:
			b.WriteString(fmt.Sprintf(`"%s"`, v))
		default:
			b.WriteString(fmt.Sprintf("%v", v))
		}
		if index != len(params) {
			b.WriteRune(',')
		}
	}
	b.WriteString("}")
	return b.String()
}

// MakeOnMatchSet 返回 ON MATCH SET 字符串。
func MakeOnMatchSet(alias string, params g.Map) string {
	if len(params) == 0 {
		return ""
	}

	b := strings.Builder{}
	index := 0
	for k, v := range params {
		index++
		b.WriteString(fmt.Sprintf("%s.%s=", alias, k))
		switch v.(type) {
		case time.Time:
			b.WriteString(fmt.Sprintf(`%v`, v.(time.Time).Unix()))
		case string:
			b.WriteString(fmt.Sprintf(`"%s"`, v))
		default:
			b.WriteString(fmt.Sprintf("%v", v))
		}
		if index != len(params) {
			b.WriteRune(',')
		}
	}
	return "ON MATCH SET " + b.String()
}
