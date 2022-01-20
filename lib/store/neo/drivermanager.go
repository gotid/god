package neo

import (
	"io"
	"sync"

	"git.zc0901.com/go/god/lib/syncx"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var driverManager = syncx.NewResourceManager()

type cachedDriver struct {
	neo4j.Driver
	once sync.Once
}

// 获取可复用的 neo4j.Driver。
func getDriver(target, username, password, realm string) (neo4j.Driver, error) {
	conn, err := getCachedDriver(target, username, password, realm)
	if err != nil {
		return nil, err
	}

	conn.once.Do(func() {
		err = conn.VerifyConnectivity()
	})
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// 从缓存池中获取 neo4j.Driver。
func getCachedDriver(target, username, password, realm string) (*cachedDriver, error) {
	conn, err := driverManager.Get(target, func() (io.Closer, error) {
		d, err := newDriver(target, username, password, realm)
		if err != nil {
			return nil, err
		}
		return &cachedDriver{Driver: d}, nil
	})
	if err != nil {
		return nil, err
	}
	return conn.(*cachedDriver), nil
}

// 返回一个新的 neo4j 连接池。
func newDriver(target, username, password, realm string) (neo4j.Driver, error) {
	d, err := neo4j.NewDriver(target, neo4j.BasicAuth(
		username,
		password,
		realm,
	))
	if err != nil {
		return nil, err
	}

	return d, nil
}
