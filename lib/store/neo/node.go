package neo

import (
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/gconv"
	jsoniter "github.com/json-iterator/go"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// ProxyNode 表示一个 neo4j 的代理节点。
type ProxyNode interface {
	// ToNeo4j 自定义节点转为 neo4j.Node。
	ToNeo4j(props interface{}) neo4j.Node
}

// Node 是一个强类型的自定义节点。
type Node struct {
	Labels []string
}

// NewNode 返回一个新的节点。
func NewNode(label ...Label) *Node {
	ls := make([]string, len(label))
	for i := 0; i < len(label); i++ {
		ls[i] = label[i].String()
	}
	return &Node{
		Labels: ls,
	}
}

var _ ProxyNode = (*Node)(nil)

// ToNeo4j 将自定义节点转为 neo4j.Node。
func (n *Node) ToNeo4j(props interface{}) neo4j.Node {
	m := n.innerProps(props)
	return neo4j.Node{
		Labels: n.Labels,
		Props:  m,
	}
}

func (n *Node) innerProps(props interface{}) g.Map {
	m := g.Map{}
	json, err := jsoniter.MarshalToString(props)
	if err != nil {
		return nil
	}

	err = jsoniter.UnmarshalFromString(json, &m)
	if err != nil {
		return nil
	}
	return m
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
