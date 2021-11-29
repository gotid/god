package neo

import (
	"fmt"
	"strings"

	"git.zc0901.com/go/god/lib/assert"
	"git.zc0901.com/go/god/lib/fx"
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/logx"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	cypherCreateNode = "CREATE (:%s $props)"
	cypherMergeNodes = `UNWIND $nodes as node
MERGE (n:%s {id:node.props.id})
ON CREATE SET n=node.props
ON MATCH SET n=node.props`
	cypherMergeNode = `MERGE (n:%s {id: $id})
%s`
	cypherDeleteNode       = `MATCH (n:%s {id: $id}) DELETE n`
	cypherDetachDeleteNode = `MATCH (n:%s {id: $id}) DETACH DELETE n`

	cypherMergeRelation = `MERGE (n1:%s {id: $id1})
MERGE (n2:%s {id: $id2})
MERGE (n1)%s(n2)
%s`
	cypherDeleteRelation = `MATCH (n1:%s {id: $id1})%s(n2:%s {id: $id2}) DELETE r`
)

// CreateNode 创建节点。
func (d *driver) CreateNode(ctx Context, nodes ...neo4j.Node) error {
	assert.IsNotNil(nodes, "节点不能为空")

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

// ReplaceNodes 批量替换节点。
func (d *driver) ReplaceNodes(ctx Context, nodes ...neo4j.Node) error {
	assert.IsNotNil(nodes, "节点不能为空")

	nodeMap := groupNodes(nodes)
	type lns struct {
		Labels string
		Nodes  []neo4j.Node
	}

	fx.From(func(source chan<- interface{}) {
		for ls, ns := range nodeMap {
			source <- lns{Labels: ls, Nodes: ns}
		}
	}).Parallel(func(item interface{}) {
		v := item.(lns)
		err := d.doGroupMerge(ctx, v.Labels, v.Nodes)
		if err != nil {
			logx.Errorf("合并失败! %v", err)
			return
		}
	})

	return nil
}

// MergeNode 合成节点并更新属性。
func (d *driver) MergeNode(ctx Context, node neo4j.Node) error {
	assert.IsNotNil(node, "节点不能为空")

	ctx.Params = g.Map{"id": node.Id}
	err := d.Run(ctx, nil, fmt.Sprintf(cypherMergeNode,
		LabelExp(node),
		MakeOnMatchSet("n", node.Props),
	))
	if err != nil {
		return err
	}

	return nil
}

// MergeRelation 合成两节点间关系。
func (d *driver) MergeRelation(ctx Context, n1 neo4j.Node, r Relation, n2 neo4j.Node) error {
	MustFullNode(n1, "n1")
	MustFullRelation(r, "r")
	MustFullNode(n2, "n2")

	ctx.Params = g.Map{"id1": n1.Id, "id2": n2.Id}
	cypher := fmt.Sprintf(cypherMergeRelation,
		LabelExp(n1), LabelExp(n2), r.Edge("r"), r.OnSet("r"))
	err := d.Run(ctx, nil, cypher)
	return err
}

// DeleteRelation 删除两节点间关系。
func (d *driver) DeleteRelation(ctx Context, n1 neo4j.Node, r Relation, n2 neo4j.Node) error {
	MustFullNode(n1, "n1")
	MustFullRelation(r, "r")
	MustFullNode(n2, "n2")

	ctx.Params = g.Map{"id1": n1.Id, "id2": n2.Id}
	cypher := fmt.Sprintf(cypherDeleteRelation, LabelExp(n1), r.Edge("r"), LabelExp(n2))
	err := d.Run(ctx, nil, cypher)
	return err
}

// DeleteNode 删除节点。
func (d *driver) DeleteNode(ctx Context, n neo4j.Node) error {
	if n.Id == 0 || len(n.Labels) == 0 {
		return nil
	}

	ctx.Params = g.Map{"id": n.Id}
	return d.Run(ctx, nil, fmt.Sprintf(cypherDeleteNode, LabelExp(n)))
}

// DetachDeleteNode 删除节点及其关系。
func (d *driver) DetachDeleteNode(ctx Context, n neo4j.Node) error {
	if n.Id == 0 || len(n.Labels) == 0 {
		return nil
	}

	ctx.Params = g.Map{"id": n.Id}
	return d.Run(ctx, nil, fmt.Sprintf(cypherDetachDeleteNode, LabelExp(n)))
}

func (d *driver) doGroupMerge(ctx Context, labels string, nodes []neo4j.Node) error {
	vs := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		node.Props["id"] = node.Id // cypher 中无法通过 node.id 获取
		vs[i] = map[string]interface{}{
			"props": node.Props,
		}
	}
	ctx.Params = g.Map{"nodes": vs}
	err := d.Run(ctx, nil, fmt.Sprintf(cypherMergeNodes, labels))
	if err != nil {
		return err
	}

	return nil
}

// SingleOtherNode 返回单边关系中另一节点。
func (d *driver) SingleOtherNode(ctx Context, input neo4j.Node, rel Relation) (neo4j.Node, error) {
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
		return neo4j.Node{}, err
	}

	return out.Node, nil
}

// GetDegree 返回指定节点全部或某个关系的度数。
func (d *driver) GetDegree(ctx Context, input neo4j.Node, rel Relation) (int64, error) {
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
func groupNodes(nodes []neo4j.Node) (ret map[string][]neo4j.Node) {
	if len(nodes) == 0 {
		return nil
	}

	ret = make(map[string][]neo4j.Node, 0)
	fx.From(func(source chan<- interface{}) {
		for _, v := range nodes {
			source <- v
		}
	}).ForEach(func(item interface{}) {
		n := item.(neo4j.Node)
		labels := LabelExp(n)
		ret[labels] = append(ret[labels], n)
	})

	return
}
