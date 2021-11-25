package sqlx

import (
	"database/sql"
	"fmt"
)

type (
	// 开启一个数据库事务
	beginTxFn func(*sql.DB) (TxSession, error)

	// TxSession 该接口表示一个数据库事务的会话。
	TxSession interface {
		Session
		Commit() error
		Rollback() error
	}

	// txSession 数据库事务结构体
	txSession struct {
		*sql.Tx
	}
)

// Query 带事务查询
func (tx txSession) Query(dest interface{}, query string, args ...interface{}) error {
	return doQuery(tx.Tx, func(rows *sql.Rows) error {
		return scan(dest, rows)
	}, query, args...)
}

// Exec 带事务执行
func (tx txSession) Exec(query string, args ...interface{}) (sql.Result, error) {
	return doExec(tx.Tx, query, args...)
}

// Prepare 带事务创建预编译语句
func (tx txSession) Prepare(query string) (StmtSession, error) {
	stmt, err := tx.Tx.Prepare(query)
	if err != nil {
		return nil, err
	}

	return statement{
		query: query,
		stmt:  stmt,
	}, nil
}

func beginTx(db *sql.DB) (TxSession, error) {
	if tx, err := db.Begin(); err != nil {
		return nil, err
	} else {
		return txSession{Tx: tx}, nil
	}
}

// doTx 执行一个事务
func doTx(c *conn, beginTx beginTxFn, transact TransactFn) (err error) {
	var db *sql.DB
	db, err = getConn(c.driverName, c.dataSourceName)
	if err != nil {
		logConnError(c.dataSourceName, err)
		return err
	}

	var tx TxSession
	tx, err = beginTx(db)
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务来自 %v, 回滚失败: %v", p, e)
			} else {
				err = fmt.Errorf("事务回滚成功，源于错误: %v", p)
			}
		} else if err != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务失败: %s, 回滚失败: %s", err, e)
			}
		} else {
			err = tx.Commit()
		}
	}()

	return transact(tx)
}
