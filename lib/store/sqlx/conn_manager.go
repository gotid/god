package sqlx

import (
	"database/sql"
	"git.zc0901.com/go/god/lib/syncx"
	"io"
	"sync"
	"time"
)

const (
	maxOpenConns = 64          // 允许最大的打开连接数
	maxIdleConns = 64          // 允许的最大空闲连接数
	maxLifetime  = time.Minute // 允许的最大连接空闲时间
)

var connManager = syncx.NewResourceManager()

// 缓存的数据库连接结构
type cachedConn struct {
	*sql.DB
	once sync.Once
}

// getConn 从缓存池中获取可复用的数据库
func getConn(driverName, dataSourceName string) (*sql.DB, error) {
	conn, err := getCachedConn(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// 尝试连接数据库（仅在第一次调用 getConn 时，下述方法才调用）
	conn.once.Do(func() {
		err = conn.Ping()
	})
	if err != nil {
		return nil, err
	}

	return conn.DB, nil
}

// getCachedConn 从缓存池中获取连接
func getCachedConn(driverName, dataSourceName string) (*cachedConn, error) {
	// 一个DSN，对应一个缓存连接
	cc, err := connManager.Get(dataSourceName, func() (io.Closer, error) {
		// 无缓存连接，则新建并通过该函数回调并加入缓存
		conn, err := newConn(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		return &cachedConn{DB: conn}, nil
	})
	if err != nil {
		return nil, err
	}
	return cc.(*cachedConn), nil
}

// newConn 新建数据库连接
func newConn(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(maxLifetime)

	return db, nil
}
