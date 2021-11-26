package neo

import (
	"fmt"
	"strings"

	"git.zc0901.com/go/god/lib/logx"

	"git.zc0901.com/go/god/lib/fx"

	"git.zc0901.com/go/god/lib/g"

	"git.zc0901.com/go/god/lib/assert"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	cypherCreateNode = "CREATE (:%s $props)"
	cypherMergeNode  = `UNWIND $nodes as node
MERGE (n:%s {id:node.props.id})
ON CREATE SET n=node.props
ON MATCH SET n=node.props`
)

// CreateNode 创建节点。
func (d *driver) CreateNode(nodes ...*neo4j.Node) error {
	assert.IsNotNil(nodes, "节点的不能为 nil")

	for _, node := range nodes {
		labels := strings.Join(node.Labels, ":")
		cypher := fmt.Sprintf(cypherCreateNode, labels)
		err := d.Run(nil, cypher, g.Map{"props": node.Props})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *driver) MergeNode(nodes ...*neo4j.Node) error {
	nodeMap := groupNodes(nodes)
	type lns struct {
		Labels string
		Nodes  []*neo4j.Node
	}

	fx.From(func(source chan<- interface{}) {
		for ls, ns := range nodeMap {
			source <- lns{Labels: ls, Nodes: ns}
		}
	}).Parallel(func(item interface{}) {
		v := item.(lns)
		err := d.doMerge(v.Labels, v.Nodes)
		if err != nil {
			logx.Errorf("合并失败! %v", err)
			return
		}
	})

	return nil
}

func (d *driver) doMerge(labels string, nodes []*neo4j.Node) error {
	vs := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		vs[i] = map[string]interface{}{
			"props": node.Props,
		}
	}
	err := d.Run(nil, fmt.Sprintf(cypherMergeNode, labels), g.Map{"nodes": vs})
	if err != nil {
		return err
	}

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

// 按节点标签分组
func groupNodes(nodes []*neo4j.Node) (ret map[string][]*neo4j.Node) {
	if len(nodes) == 0 {
		return nil
	}
	ret = make(map[string][]*neo4j.Node, 0)

	fx.From(func(source chan<- interface{}) {
		for _, v := range nodes {
			source <- v
		}
	}).ForEach(func(item interface{}) {
		n := item.(*neo4j.Node)
		labels := strings.Join(n.Labels, ":")
		ret[labels] = append(ret[labels], n)
	})

	return
}
