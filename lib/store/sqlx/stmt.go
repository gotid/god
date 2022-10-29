package sqlx

import (
	"context"
	"database/sql"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/timex"
	"time"
)

const defaultSlowThreshold = 500 * time.Millisecond

var (
	logSQL        = syncx.ForAtomicBool(true)
	logSlowSQL    = syncx.ForAtomicBool(true)
	slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)
)

// DisableLog 禁用 SQL 语句及慢查询 日志。
func DisableLog() {
	logSQL.Set(false)
	logSlowSQL.Set(false)
}

// DisableStmtLog 禁用 SQL 语句日志，但是保留慢查询日志。
func DisableStmtLog() {
	logSQL.Set(false)
}

// SetSlowThreshold 设置 SQL 慢查询阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}

func exec(ctx context.Context, conn sessionConn, query string, args ...interface{}) (sql.Result, error) {
	guard := newGuard("exec")
	if err := guard.start(query, args...); err != nil {
		return nil, err
	}

	result, err := conn.ExecContext(ctx, query, args...)
	guard.finish(ctx, err)

	return result, err
}

func execStmt(ctx context.Context, conn stmtConn, query string, args ...interface{}) (sql.Result, error) {
	guard := newGuard("execStmt")
	if err := guard.start(query, args...); err != nil {
		return nil, err
	}

	result, err := conn.ExecContext(ctx, args...)
	guard.finish(ctx, err)

	return result, err
}

func query(ctx context.Context, conn sessionConn, scanner func(*sql.Rows) error, query string, args ...interface{}) error {
	guard := newGuard("query")
	if err := guard.start(query, args...); err != nil {
		return err
	}

	rows, err := conn.QueryContext(ctx, query, args...)
	guard.finish(ctx, err)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanner(rows)
}

func queryStmt(ctx context.Context, conn stmtConn, scanner func(*sql.Rows) error, query string, args ...interface{}) error {
	guard := newGuard("queryStmt")
	if err := guard.start(query, args...); err != nil {
		return err
	}

	rows, err := conn.QueryContext(ctx, args...)
	guard.finish(ctx, err)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanner(rows)
}

type (
	sqlGuard interface {
		start(query string, args ...interface{}) error
		finish(ctx context.Context, err error)
	}

	nilGuard struct{}

	realSqlGuard struct {
		command   string
		stmt      string
		startTime time.Duration
	}
)

func (n nilGuard) start(_ string, _ ...interface{}) error {
	return nil
}

func (n nilGuard) finish(_ context.Context, _ error) {
}

func (g *realSqlGuard) start(query string, args ...interface{}) error {
	stmt, err := format(query, args...)
	if err != nil {
		return err
	}

	g.stmt = stmt
	g.startTime = timex.Now()

	return nil
}

func (g *realSqlGuard) finish(ctx context.Context, err error) {
	duration := timex.Since(g.startTime)
	if duration > slowThreshold.Load() {
		logx.WithContext(ctx).WithDuration(duration).Slowf("[SQL] %s：慢查询 - %s", g.command, g.stmt)
	} else if logSQL.True() {
		logx.WithContext(ctx).WithDuration(duration).Infof("sql %s: %s", g.command, g.stmt)
	}

	if err != nil {
		logSQLError(ctx, g.stmt, err)
	}

	metricReqDur.Observe(int64(duration/time.Millisecond), g.command)
}

func newGuard(command string) sqlGuard {
	if logSQL.True() || logSlowSQL.True() {
		return &realSqlGuard{command: command}
	}

	return nilGuard{}
}
