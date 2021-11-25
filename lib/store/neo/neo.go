package neo

import (
	"git.zc0901.com/go/god/lib/g"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// Driver 表示一个带有断路器保护的 neo4j 驱动。
type Driver interface {
	Session
}

// TransactFn 事务型执行函数
type TransactFn func(tx neo4j.Transaction) error

// Session 表示一个可进行 neo4j 读写的会话。
type Session interface {
	// Driver 返回一个复用的 neo4j.Driver。
	Driver() (neo4j.Driver, error)
	// BeginTx 返回一个新的事务。
	BeginTx() (neo4j.Transaction, error)
	// Transact 执行事务型操作。
	Transact(fn TransactFn) error
	// Read 读数 —— 运行Cypher并读入目标。
	Read(dest interface{}, cypher string, params ...g.Map) error
	// TxRead 事务型读数 —— 运行Cypher并读入目标。
	TxRead(tx neo4j.Transaction, dest interface{}, cypher string, params ...g.Map) error
	// Run 运行 —— 并利用扫描器扫描指定Cypher的执行结果。
	Run(scanner Scanner, cypher string, params ...g.Map) error
	// TxRun 事务型运行 —— 利用扫描器扫描指定Cypher的执行结果。
	TxRun(tx neo4j.Transaction, scanner Scanner, cypher string, params ...g.Map) error
}

// NewNeo 返回新的 Neo 驱动。出错则退出。
func NewNeo(target, username, password, realm string) Driver {
	return MustDriver(target, username, password, realm)
}
