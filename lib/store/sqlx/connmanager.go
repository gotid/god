package sqlx

import (
	"database/sql"
	"github.com/gotid/god/lib/syncx"
	"io"
	"sync"
	"time"
)

const (
	maxIdleConns = 64
	maxOpenConns = 64
	maxLifetime  = time.Minute
)

var connManager = syncx.NewResourceManager()

type pingedDB struct {
	*sql.DB
	once sync.Once
}

func getConn(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := getCachedDB(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	db.once.Do(func() {
		err = db.Ping()
	})
	if err != nil {
		return nil, err
	}

	return db.DB, nil
}

func getCachedDB(driverName string, dataSourceName string) (*pingedDB, error) {
	val, err := connManager.Get(dataSourceName, func() (io.Closer, error) {
		db, err := newDB(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}

		return &pingedDB{
			DB: db,
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return val.(*pingedDB), nil
}

func newDB(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	// 受 https://github.com/golang/go/issues/9851
	// 和 https://github.com/go-sql-driver/mysql/issues/257
	// 影响，当前需进行如下操作。
	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(maxLifetime)

	return db, nil
}
