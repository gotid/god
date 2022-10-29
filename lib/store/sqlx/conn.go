package sqlx

import (
	"context"
	"database/sql"
	"github.com/gotid/god/lib/breaker"
	"github.com/gotid/god/lib/logx"
)

// ErrNotFound 是 sql.ErrNoRows 的别名。
var ErrNotFound = sql.ErrNoRows

type (
	// Session 接口表示一个原始数据库链接或事务会话。
	Session interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		ExecCtx(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		Prepare(query string) (StmtSession, error)
		PrepareCtx(ctx context.Context, query string) (StmtSession, error)
		QueryRow(v interface{}, query string, args ...interface{}) error
		QueryRowCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error
		QueryRowPartial(v interface{}, query string, args ...interface{}) error
		QueryRowPartialCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error
		QueryRows(v interface{}, query string, args ...interface{}) error
		QueryRowsCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error
		QueryRowsPartial(v interface{}, query string, args ...interface{}) error
		QueryRowsPartialCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error
	}

	// StmtSession 接口表示一个用于执行语句的会话。
	StmtSession interface {
		Close() error
		Exec(args ...interface{}) (sql.Result, error)
		ExecCtx(ctx context.Context, args ...interface{}) (sql.Result, error)
		QueryRow(v interface{}, args ...interface{}) error
		QueryRowCtx(ctx context.Context, v interface{}, args ...interface{}) error
		QueryRowPartial(v interface{}, args ...interface{}) error
		QueryRowPartialCtx(ctx context.Context, v interface{}, args ...interface{}) error
		QueryRows(v interface{}, args ...interface{}) error
		QueryRowsCtx(ctx context.Context, v interface{}, args ...interface{}) error
		QueryRowsPartial(v interface{}, args ...interface{}) error
		QueryRowsPartialCtx(ctx context.Context, v interface{}, args ...interface{}) error
	}

	// Conn 接口代表原始连接，封装事务方法 Transact。
	Conn interface {
		Session
		RawDB() (*sql.DB, error) // 供其他 ORM 操作的原始连接，请勿关闭，小心使用。
		Transact(fn func(Session) error) error
		TransactCtx(ctx context.Context, fn func(context.Context, Session) error) error
	}

	// Option 自定义一个 sql 连接的方法。
	Option func(*commonConn)

	// 线程安全的通用数据库连接。
	// 因 CORBA 不支持 PREPARE，故合并 query 参数为一个字符串并进行底层无参 query。
	commonConn struct {
		brk      breaker.Breaker
		provider connProvider
		onError  func(err error)
		accept   func(error) bool
		beginTx  beginnable
	}

	sessionConn interface {
		Exec(query string, args ...interface{}) (sql.Result, error)
		ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	}

	stmtConn interface {
		Exec(args ...interface{}) (sql.Result, error)
		ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
		Query(args ...interface{}) (*sql.Rows, error)
		QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error)
	}

	connProvider func() (*sql.DB, error)

	statement struct {
		query string
		stmt  *sql.Stmt
	}
)

// NewConn 返回一个给定驱动和数据源的 Conn 数据库连接。
func NewConn(driverName, dataSourceName string, opts ...Option) Conn {
	conn := &commonConn{
		brk: breaker.New(),
		provider: func() (*sql.DB, error) {
			return getConn(driverName, dataSourceName)
		},
		onError: func(err error) {
			logInstanceError(dataSourceName, err)
		},
		beginTx: begin,
	}
	for _, opt := range opts {
		opt(conn)
	}

	return conn
}

// NewConnFromDB 返回给定 sql.DB 的原始连接 Conn。
// ，用于其他 ORM 交互，小心使用。
func NewConnFromDB(db *sql.DB, opts ...Option) Conn {
	conn := &commonConn{
		brk: breaker.New(),
		provider: func() (*sql.DB, error) {
			return db, nil
		},
		onError: func(err error) {
			logx.Errorf("获取 SQL 实例错误：%v", err)
		},
		beginTx: begin,
	}

	for _, opt := range opts {
		opt(conn)
	}

	return conn
}

func (db *commonConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.ExecCtx(context.Background(), query, args...)
}

func (db *commonConn) ExecCtx(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	ctx, span := startSpan(ctx, "Exec")
	defer func() {
		endSpan(span, err)
	}()

	err = db.brk.DoWithAcceptable(func() error {
		var conn *sql.DB
		conn, err = db.provider()
		if err != nil {
			db.onError(err)
			return err
		}

		result, err = exec(ctx, conn, query, args...)
		return err
	}, db.acceptable)
	if err == breaker.ErrServiceUnavailable {
		metricReqErr.Inc("Exec", "breaker")
	}

	return
}

func (db *commonConn) Prepare(query string) (StmtSession, error) {
	return db.PrepareCtx(context.Background(), query)
}

func (db *commonConn) PrepareCtx(ctx context.Context, query string) (stmt StmtSession, err error) {
	ctx, span := startSpan(ctx, "Prepare")
	defer func() {
		endSpan(span, err)
	}()

	err = db.brk.DoWithAcceptable(func() error {
		var conn *sql.DB
		conn, err = db.provider()
		if err != nil {
			db.onError(err)
			return err
		}

		st, err := conn.PrepareContext(ctx, query)
		if err != nil {
			return err
		}

		stmt = statement{
			query: query,
			stmt:  st,
		}

		return nil
	}, db.acceptable)
	if err == breaker.ErrServiceUnavailable {
		metricReqErr.Inc("Prepare", "breaker")
	}

	return
}

func (db *commonConn) QueryRow(v interface{}, query string, args ...interface{}) error {
	return db.QueryRowCtx(context.Background(), v, query, args...)
}

func (db *commonConn) QueryRowCtx(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRow")
	defer func() {
		endSpan(span, err)
	}()

	return db.queryRows(ctx, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, true)
	}, query, args...)
}

func (db *commonConn) QueryRowPartial(v interface{}, query string, args ...interface{}) error {
	return db.QueryRowPartialCtx(context.Background(), v, query, args...)
}

func (db *commonConn) QueryRowPartialCtx(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRowPartial")
	defer func() {
		endSpan(span, err)
	}()

	return db.queryRows(ctx, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, false)
	}, query, args...)
}

func (db *commonConn) QueryRows(v interface{}, query string, args ...interface{}) error {
	return db.QueryRowsCtx(context.Background(), v, query, args...)
}

func (db *commonConn) QueryRowsCtx(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRows")
	defer func() {
		endSpan(span, err)
	}()

	return db.queryRows(ctx, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, true)
	}, query, args...)
}

func (db *commonConn) QueryRowsPartial(v interface{}, query string, args ...interface{}) error {
	return db.QueryRowsPartialCtx(context.Background(), v, query, args...)
}

func (db *commonConn) QueryRowsPartialCtx(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRowsPartial")
	defer func() {
		endSpan(span, err)
	}()

	return db.queryRows(ctx, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, false)
	}, query, args...)
}

func (db *commonConn) RawDB() (*sql.DB, error) {
	return db.provider()
}

func (db *commonConn) Transact(fn func(Session) error) error {
	return db.TransactCtx(context.Background(), func(_ context.Context, session Session) error {
		return fn(session)
	})
}

func (db *commonConn) TransactCtx(ctx context.Context, fn func(context.Context, Session) error) (err error) {
	ctx, span := startSpan(ctx, "Transact")
	defer func() {
		endSpan(span, err)
	}()

	err = db.brk.DoWithAcceptable(func() error {
		return transact(ctx, db, db.beginTx, fn)
	}, db.acceptable)
	if err == breaker.ErrServiceUnavailable {
		metricReqErr.Inc("Transact", "breaker")
	}

	return
}

func (db *commonConn) acceptable(err error) bool {
	ok := err == nil || err == sql.ErrNoRows || err == sql.ErrTxDone || err == context.Canceled
	if db.accept == nil {
		return ok
	}

	return ok || db.accept(err)
}

func (db *commonConn) queryRows(ctx context.Context, scanner func(*sql.Rows) error, q string, args ...interface{}) (err error) {
	var scanErr error
	err = db.brk.DoWithAcceptable(func() error {
		conn, e := db.provider()
		if e != nil {
			db.onError(e)
			return e
		}

		return query(ctx, conn, func(rows *sql.Rows) error {
			scanErr = scanner(rows)
			return scanErr
		}, q, args...)
	}, func(err error) bool {
		return scanErr == err || db.acceptable(err)
	})
	if err == breaker.ErrServiceUnavailable {
		metricReqErr.Inc("queryRows", "breaker")
	}

	return
}

func (s statement) Close() error {
	return s.stmt.Close()
}

func (s statement) Exec(args ...interface{}) (sql.Result, error) {
	return s.ExecCtx(context.Background(), args...)
}

func (s statement) ExecCtx(ctx context.Context, args ...interface{}) (result sql.Result, err error) {
	ctx, span := startSpan(ctx, "Exec")
	defer func() {
		endSpan(span, err)
	}()

	return execStmt(ctx, s.stmt, s.query, args...)
}

func (s statement) QueryRow(v interface{}, args ...interface{}) error {
	return s.QueryRowCtx(context.Background(), v, args...)
}

func (s statement) QueryRowCtx(ctx context.Context, v interface{}, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRow")
	defer func() {
		endSpan(span, err)
	}()

	return queryStmt(ctx, s.stmt, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, true)
	}, s.query, args...)
}

func (s statement) QueryRowPartial(v interface{}, args ...interface{}) error {
	return s.QueryRowPartialCtx(context.Background(), v, args...)
}

func (s statement) QueryRowPartialCtx(ctx context.Context, v interface{}, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRowPartial")
	defer func() {
		endSpan(span, err)
	}()

	return queryStmt(ctx, s.stmt, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, false)
	}, s.query, args...)
}

func (s statement) QueryRows(v interface{}, args ...interface{}) error {
	return s.QueryRowsCtx(context.Background(), v, args...)
}

func (s statement) QueryRowsCtx(ctx context.Context, v interface{}, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRows")
	defer func() {
		endSpan(span, err)
	}()

	return queryStmt(ctx, s.stmt, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, true)
	}, s.query, args...)
}

func (s statement) QueryRowsPartial(v interface{}, args ...interface{}) error {
	return s.QueryRowsPartialCtx(context.Background(), v, args...)
}

func (s statement) QueryRowsPartialCtx(ctx context.Context, v interface{}, args ...interface{}) (err error) {
	ctx, span := startSpan(ctx, "QueryRowsPartial")
	defer func() {
		endSpan(span, err)
	}()

	return queryStmt(ctx, s.stmt, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, false)
	}, s.query, args...)
}
