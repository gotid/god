package neo

import (
	"fmt"
	"strings"

	"git.zc0901.com/go/god/lib/g"

	"git.zc0901.com/go/god/lib/logx"

	"git.zc0901.com/go/god/lib/fx"

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
func (d *driver) CreateNode(ctx Context, nodes ...*neo4j.Node) error {
	assert.IsNotNil(nodes, "节点的不能为 nil")

	for _, node := range nodes {
		labels := strings.Join(node.Labels, ":")
		cypher := fmt.Sprintf(cypherCreateNode, labels)
		ctx.Params = g.Map{"props": node.Props}
		err := d.Run(ctx, nil, cypher)
		if err != nil {
			return err
		}
	}

	return nil
}

// MergeNode 合并节点。
func (d *driver) MergeNode(ctx Context, nodes ...*neo4j.Node) error {
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
		err := d.doMerge(ctx, v.Labels, v.Nodes)
		if err != nil {
			logx.Errorf("合并失败! %v", err)
			return
		}
	})

	return nil
}

func (d *driver) doMerge(ctx Context, labels string, nodes []*neo4j.Node) error {
	vs := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		vs[i] = map[string]interface{}{
			"props": node.Props,
		}
	}
	ctx.Params = g.Map{"nodes": vs}
	err := d.Run(ctx, nil, fmt.Sprintf(cypherMergeNode, labels))
	if err != nil {
		return err
	}

	return nil
}

// SingleOtherNode 返回单一关系中的另一节点。
func (d *driver) SingleOtherNode(ctx Context, input *neo4j.Node, rel *Relationship) (*neo4j.Node, error) {
	assert.IsNotNil(input, "节点不可为空")
	assert.IsNotEmpty(input.Id, "节点编号不可为0")
	assert.IsNotNil(rel, "单一节点的关系必须明确")
	assert.IsNotEmpty(rel.Type, "单一阶段的关系类型必须明确")

	var out struct {
		Node neo4j.Node `neo:"o"`
	}
	cypher := fmt.Sprintf(`MATCH (i)%s(o) WHERE i.id=$id RETURN o`, rel.Edge())
	ctx.Params = g.Map{"id": input.Id}
	err := d.Read(ctx, &out, cypher)
	if err != nil {
		return nil, err
	}

	return &out.Node, nil
}

// GetDegree 返回指定节点全部或某一中关系的 Degree 数量
func (d *driver) GetDegree(ctx Context, input *neo4j.Node, rel *Relationship) (int64, error) {
	assert.IsNotNil(input, "节点不可为空")
	assert.IsNotEmpty(input.Id, "节点编号不可为0")

	var degree int64
	cypher := fmt.Sprintf(
		`MATCH (i)%s() WHERE i.id=$id RETURN COUNT(r) as degree`,
		rel.Edge("r"),
	)
	ctx.Params = g.Map{"id": input.Id}
	err := d.Read(ctx, &degree, cypher)
	if err != nil {
		return 0, err
	}

	return degree, nil
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
