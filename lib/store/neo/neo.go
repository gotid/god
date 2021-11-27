package neo

import (
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/gconv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// Driver 表示一个带有断路器保护的 neo4j 驱动。
type Driver interface {
	Session
}

// Session 表示一个可进行 neo4j 读写的会话。
type Session interface {
	// Driver 返回一个复用的 neo4j.Driver。
	Driver() neo4j.Driver
	// BeginTx 返回一个新的事务。
	BeginTx() (neo4j.Transaction, error)
	// Transact 执行事务型操作。
	Transact(fn TransactFn) error
	// Read 读数 —— 运行Cypher并读入目标。
	Read(ctx Context, dest interface{}, cypher string) error
	// Run 运行 —— 并利用扫描器扫描指定Cypher的执行结果。
	Run(ctx Context, scanner Scanner, cypher string) error

	// CreateNode 创建节点。
	CreateNode(ctx Context, nodes ...neo4j.Node) error
	// MergeNode 合成节点并覆盖属性。
	MergeNode(ctx Context, nodes ...neo4j.Node) error
	// DetachNode 删除节点及其关系。
	DetachNode(ctx Context, n neo4j.Node) error
	// Relate 合成两节点间关系。
	Relate(ctx Context, n1 neo4j.Node, r Relation, n2 neo4j.Node) error
	// SingleOtherNode 返回单边关系中另一节点。
	SingleOtherNode(ctx Context, input neo4j.Node, rel Relation) (neo4j.Node, error)
	// GetDegree 返回指定节点的 Degree 数量
	GetDegree(ctx Context, input neo4j.Node, rel Relation) (int64, error)
}

// NewNeo 返回新的 Neo 驱动。出错则退出。
func NewNeo(target, username, password, realm string) Driver {
	return MustDriver(target, username, password, realm)
}

// TransactFn 事务型执行函数
type TransactFn func(tx neo4j.Transaction) error

// Context 是一个驱动执行参数
type Context struct {
	Tx     neo4j.Transaction
	Params g.Map
	Driver neo4j.Driver
}

// Map 返回一个没有事务的映射参数
func Map(kv ...interface{}) Context {
	if len(kv)%2 != 0 {
		panic("kv对必须为偶数")
	}

	m := g.Map{}

	for i := 0; i < len(kv)/2; i++ {
		m[gconv.String(kv[i])] = kv[i+1]
	}

	return Context{Params: m}
}
