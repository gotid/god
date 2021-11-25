package neo

import (
	"fmt"

	"git.zc0901.com/go/god/lib/g"

	"git.zc0901.com/go/god/lib/assert"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// CreateNode 创建一个节点。
func (d *driver) CreateNode(node *neo4j.Node) error {
	assert.IsNotNil(node, "节点的不能为 nil")
	assert.IsNotEmpty(node.Id, "节点Id不能为0")

	fmt.Println(node)

	return nil
}

// SingleOtherNode 返回单一关系中的另一节点。
func (d *driver) SingleOtherNode(input *neo4j.Node, rel *Relationship) (*neo4j.Node, error) {
	assert.IsNotNil(rel, "单一节点的关系必须明确")
	assert.IsNotEmpty(rel.Type, "单一阶段的关系类型必须明确")

	var out neo4j.Node
	cypher := fmt.Sprintf(`MATCH (i)%s(o) WHERE id(i)=$id RETURN o`, rel.Edge())
	err := d.Read(&out, cypher, g.Map{"id": input.Id})
	if err != nil {
		return nil, err
	}

	return &out, nil
}
