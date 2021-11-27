package neo

import (
	"strings"

	"git.zc0901.com/go/god/lib/assert"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// Labels 返回 neo4j.Node 标签切片的特征字符串。
func Labels(n neo4j.Node) string {
	return strings.Join(n.Labels, ":")
}

// MustFullNode 必须为完整节点，否则退出程序。
func MustFullNode(node neo4j.Node, name string) {
	assert.IsNotNil(node, name, "节点不能为空")
	assert.IsNotEmpty(node.Id, name, "Id 不能为空")
	assert.IsNotEmpty(Labels(node), name, "标签不能为空")
}

// MustFullRelation 必须为完整关系，否则退出程序。
func MustFullRelation(r Relation, name string) {
	assert.IsNotEmpty(r, name, "关系不能为空")
	assert.IsNotEmpty(r.Type, name, "关系类型必须明确")
	assert.IsNotEmpty(r.Direction, name, "关系方向必须明确")
}
