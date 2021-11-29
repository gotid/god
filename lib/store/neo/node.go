package neo

import (
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/gconv"
	jsoniter "github.com/json-iterator/go"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// ProxyNode 表示一个 neo4j 的代理节点。
type ProxyNode interface {
	// InnerProps 强类型属性转 g.Map
	InnerProps() g.Map

	// ToNeo4j 自定义节点转为 neo4j.Node。
	ToNeo4j(g.Map) neo4j.Node
}

// Node 是一个强类型的自定义节点。
type Node struct {
	// ProxyNode
	Labels []string
}

var _ ProxyNode = (*Node)(nil)

func (n *Node) InnerProps() g.Map {
	// TODO implement me
	panic("implement me by sub struct")
}

// ToNeo4j 将自定义节点转为 neo4j.Node。
func (n *Node) ToNeo4j(props g.Map) neo4j.Node {
	return neo4j.Node{
		Labels: n.Labels,
		Props:  props,
	}
}

// ConvNode 从 neo4j.Node 转为自定义结构体。
func ConvNode(source neo4j.Node, dest interface{}) (err error) {
	err = gconv.Struct(source, dest)
	if err != nil {
		return err
	}

	return
}

// ConvNodes 从 []neo4j.Node 转为自定义结构体切片组。
func ConvNodes(source []neo4j.Node, dest interface{}) (err error) {
	err = gconv.Structs(source, dest)
	if err != nil {
		return err
	}

	return
}

// ProxyProps 表示一个 neo4j 的代理属性。
type ProxyProps interface {
	// InnerProps 强类型属性转 g.Map
	InnerProps(prop interface{}) g.Map
}

// Props 是一个强类型的基础属性。
type Props struct {
	Id int64 `json:"id"`
}

var _ ProxyProps = (*Props)(nil)

// InnerProps 强类型属性转 g.Map
func (p *Props) InnerProps(prop interface{}) g.Map {
	m := g.Map{}
	json, err := jsoniter.MarshalToString(prop)
	if err != nil {
		return nil
	}

	err = jsoniter.UnmarshalFromString(json, &m)
	if err != nil {
		return nil
	}
	return m
}
