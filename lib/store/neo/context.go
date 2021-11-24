package neo

import (
	"git.zc0901.com/go/god/lib/g"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// Context 是一个 neo 上下文。
type Context struct {
	config *Config
	driver neo4j.Driver
}

func MustContext(c *Config) *Context {
	driver, err := neo4j.NewDriver(c.Target, neo4j.BasicAuth(
		c.Username,
		c.Password,
		"",
	))
	if err != nil {
		panic(err)
	}

	return &Context{
		config: c,
		driver: driver,
	}
}

// Config 返回配置信息。
func (c *Context) Config() *Config {
	return c.config
}

// Driver 返回 neo4j 连接池。
func (c *Context) Driver() neo4j.Driver {
	return c.driver
}

// BeginTx 开启并返回一个事务。
func (c *Context) BeginTx() (neo4j.Transaction, error) {
	session := c.driver.NewSession(neo4j.SessionConfig{})
	tx, err := session.BeginTransaction()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// RunCypher 运行指定 Cypher 并进行回调处理。
func (c *Context) RunCypher(cypher string, scanner Scanner,
	params ...g.Map) error {
	if scanner == nil {
		return ErrNoScanner
	}
	var param g.Map
	if len(params) > 0 {
		param = params[0]
	}

	session := c.driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	result, err := session.Run(cypher, param)
	if err != nil {
		return err
	}

	return scanner(result)
}

// RunCypherWithTx 在事务中运行指定 Cypher 并进行回调处理。
func (c *Context) RunCypherWithTx(tx neo4j.Transaction, cypher string,
	callback Scanner, params ...g.Map) error {
	if callback == nil {
		return ErrNoScanner
	}
	var param g.Map
	if len(params) > 0 {
		param = params[0]
	}

	result, err := tx.Run(cypher, param)
	if err != nil {
		return err
	}

	return callback(result)
}
