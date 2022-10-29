package sqlx

import (
	"context"
	"database/sql"
	"fmt"
)

type (
	trans interface {
		Session
		Commit() error
		Rollback() error
	}

	beginnable func(db *sql.DB) (trans, error)

	txSession struct {
		*sql.Tx
	}
)

// NewSessionFromTx 返回给定事务 tx 对应的会话 Session。
func NewSessionFromTx(tx *sql.Tx) Session {
	return txSession{Tx: tx}
}

func (t txSession) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.ExecCtx(context.Background(), query, args...)
}

func (t txSession) ExecCtx(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	ctx, span := startSpan(ctx, "Exec")
	defer func() {
		endSpan(span, err)
	}()

	result, err = exec(ctx, t.Tx, query, args...)

	return
}

func (t txSession) Prepare(query string) (StmtSession, error) {
	return t.PrepareCtx(context.Background(), query)
}

func (t txSession) PrepareCtx(ctx context.Context, query string) (stmtSession StmtSession, err error) {
	ctx, span := startSpan(ctx, "Prepare")
	defer func() {
		endSpan(span, err)
	}()

	stmt, err := t.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return statement{
		query: query,
		stmt:  stmt,
	}, nil
}

func (t txSession) QueryRow(v interface{}, query string, args ...interface{}) error {
	return t.QueryRowCtx(context.Background(), v, query, args...)
}

func (t txSession) QueryRowCtx(ctx context.Context, v interface{}, q string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRow")
	defer func() {
		endSpan(span, err)
	}()

	return query(ctx, t.Tx, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, true)
	}, q, args...)
}

func (t txSession) QueryRowPartial(v interface{}, query string, args ...interface{}) error {
	return t.QueryRowPartialCtx(context.Background(), v, query, args...)
}

func (t txSession) QueryRowPartialCtx(ctx context.Context, v interface{}, q string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRowPartial")
	defer func() {
		endSpan(span, err)
	}()

	return query(ctx, t.Tx, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, false)
	}, q, args...)
}

func (t txSession) QueryRows(v interface{}, query string, args ...interface{}) error {
	return t.QueryRowsCtx(context.Background(), v, query, args...)
}

func (t txSession) QueryRowsCtx(ctx context.Context, v interface{}, q string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRows")
	defer func() {
		endSpan(span, err)
	}()

	return query(ctx, t.Tx, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, true)
	}, q, args...)
}

func (t txSession) QueryRowsPartial(v interface{}, query string, args ...interface{}) error {
	return t.QueryRowsPartialCtx(context.Background(), v, query, args...)
}

func (t txSession) QueryRowsPartialCtx(ctx context.Context, v interface{}, q string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRowsPartial")
	defer func() {
		endSpan(span, err)
	}()

	return query(ctx, t.Tx, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, false)
	}, q, args...)
}

func begin(db *sql.DB) (trans, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return txSession{
		Tx: tx,
	}, nil
}

func transact(ctx context.Context, db *commonConn, b beginnable, fn func(context.Context, Session) error) (err error) {
	conn, err := db.provider()
	if err != nil {
		db.onError(err)
		return err
	}

	return transactOnConn(ctx, conn, b, fn)
}

func transactOnConn(ctx context.Context, conn *sql.DB, b beginnable, fn func(context.Context, Session) error) (err error) {
	var tx trans
	tx, err = b(conn)
	if err != nil {
		return
	}

	defer func() {
		if p := recover(); p != nil {

		} else if err != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务失败了：%s，回滚也失败了：%w", err, e)
			}
		} else {
			err = tx.Commit()
		}
	}()

	return fn(ctx, tx)
}
