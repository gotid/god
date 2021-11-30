package neo

import (
	"errors"
	"reflect"

	"git.zc0901.com/go/god/lib/gconv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// ProxyNode 表示一个 neo4j 的代理节点。
type ProxyNode interface {
	// ToNeo4j 自定义节点转为 neo4j.Node。
	ToNeo4j(props interface{}, excludeKeys ...string) neo4j.Node
}

// Node 是一个强类型的自定义节点。
type Node struct {
	Id     int64 // Props["id"] 的快捷方式
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
func (n *Node) ToNeo4j(props interface{}, excludeKeys ...string) neo4j.Node {
	m := gconv.Map(props)
	for _, key := range excludeKeys {
		delete(m, key)
	}
	id := int64(0)
	if v, ok := m["id"]; ok {
		id = v.(int64)
	}
	return neo4j.Node{
		Id:     id,
		Labels: n.Labels,
		Props:  m,
	}
}

// ConvNode 从 neo4j.Node 转为自定义结构体。
func ConvNode(source neo4j.Node, dest interface{}) (err error) {
	err = gconv.Struct(source, dest)
	if err != nil {
		return err
	}

	// 重置 NodeId 为 PropsId
	return resetId(dest)
}

// 重置目标结构体中 NodeId = PropsId
func resetId(dest interface{}) error {
	dve := reflect.ValueOf(dest).Elem()

	if dve.Kind() != reflect.Struct {
		return errors.New("转换目标不是结构体")
	}

	nodeField := dve.FieldByName("Node")
	if !nodeField.IsValid() {
		return errors.New("目标结构体未包含 Node 子结构")
	}
	nodeId := nodeField.FieldByName("Id")
	if !nodeId.IsValid() {
		return errors.New("目标结构体 Node Id 无效")
	}

	propsField := dve.FieldByName("Props")
	if !propsField.IsValid() {
		return errors.New("目标结构体未包含 Props 子结构")
	}
	propsId := propsField.FieldByName("Id")
	if !propsId.IsValid() {
		return errors.New("目标结构体 Props Id 无效")
	}

	if nodeId.Kind() != propsId.Kind() {
		return errors.New("目标结构体 Props Id 和 Node Id 类型不同")
	}

	nodeId.Set(propsId)
	return nil
}
