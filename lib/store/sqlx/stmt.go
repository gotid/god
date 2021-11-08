package sqlx

import (
	"database/sql"
	"time"

	"git.zc0901.com/go/god/lib/syncx"

	"git.zc0901.com/go/god/lib/timex"

	"git.zc0901.com/go/god/lib/logx"
)

type (
	// StmtSession 该接口表示一个可以执行预编译语句的会话。
	StmtSession interface {
		Close() error
		Exec(args ...interface{}) (sql.Result, error)
		Query(dest interface{}, args ...interface{}) error
	}

	stmtConn interface {
		Exec(args ...interface{}) (sql.Result, error)
		Query(args ...interface{}) (*sql.Rows, error)
	}

	// 封装内部使用的预编译语句
	statement struct {
		query string
		stmt  *sql.Stmt
	}
)

var slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)

// SetSlowThreshold 设置慢查询时间阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}

func (s statement) Close() error {
	return s.stmt.Close()
}

func (s statement) Exec(args ...interface{}) (sql.Result, error) {
	return doExecStmt(s.stmt, s.query, args...)
}

func (s statement) Query(dest interface{}, args ...interface{}) error {
	return doQueryStmt(s.stmt, func(rows *sql.Rows) error {
		return scan(dest, rows)
	}, s.query, args...)
}

// doQueryStmt 执行预编译查询语句
func doQueryStmt(conn stmtConn, scanner func(rows *sql.Rows) error, query string, args ...interface{}) error {
	stmt, err := format(query, args...)
	if err != nil {
		return err
	}

	startTime := timex.Now()
	rows, err := conn.Query(args...)
	duration := timex.Since(startTime)
	if duration > slowThreshold.Load() {
		logx.WithDuration(duration).Slowf("[SQL] doExecStmt: 慢查询 —— %s", stmt)
	} else {
		logx.WithDuration(duration).Infof("[SQL] doExecStmt: %s", stmt)
	}
	if err != nil {
		logSqlError(stmt, err)
	}
	defer rows.Close()

	return scanner(rows)
}

// doExecStmt 执行预编译语句
func doExecStmt(conn stmtConn, query string, args ...interface{}) (sql.Result, error) {
	stmt, err := format(query, args...)
	if err != nil {
		return nil, err
	}

	startTime := timex.Now()
	result, err := conn.Exec(args...)
	duration := timex.Since(startTime)
	if duration > defaultSlowThreshold {
		logx.WithDuration(duration).Slowf("[SQL] doExecStmt: 慢查询 —— %s", stmt)
	} else {
		logx.WithDuration(duration).Infof("[SQL] doExecStmt: %s", stmt)
	}
	if err != nil {
		logSqlError(stmt, err)
	}

	return result, err
}

// doQuery 执行查询语句
func doQuery(db session, scanner func(*sql.Rows) error, query string, args ...interface{}) error {
	// 格式化后的查询字符串
	stmt, err := format(query, args...)
	if err != nil {
		return err
	}

	// 带有慢查询检测
	startTime := time.Now()
	rows, err := db.Query(query, args...)
	duration := time.Since(startTime)

	if duration > defaultSlowThreshold {
		logx.WithDuration(duration).Slowf("[SQL] 慢查询 - %s", stmt)
	} else {
		logx.WithDuration(duration).Infof("[SQL] 查询: %s", stmt)
	}

	if err != nil {
		logSqlError(stmt, err)
		return err
	}

	// 关闭数据库连接，释放资源
	defer func() {
		_ = rows.Close()
	}()

	return scanner(rows)
}

// 执行语句
func doExec(db session, query string, args ...interface{}) (sql.Result, error) {
	// 格式化后的查询字符串
	stmt, err := format(query, args...)
	if err != nil {
		return nil, err
	}

	// 转换自定义 sqlx.NullXxx 为 sql.NullXxx
	var nvArgs []interface{}
	for _, arg := range args {
		switch v := arg.(type) {
		case NullTime:
			nvArgs = append(nvArgs, sql.NullTime{Valid: v.Valid, Time: v.Time})
		case NullBool:
			nvArgs = append(nvArgs, sql.NullBool{Valid: v.Valid, Bool: v.Bool})
		case NullString:
			nvArgs = append(nvArgs, sql.NullString{Valid: v.Valid, String: v.String})
		case NullInt64:
			nvArgs = append(nvArgs, sql.NullInt64{Valid: v.Valid, Int64: v.Int64})
		case NullFloat64:
			nvArgs = append(nvArgs, sql.NullFloat64{Valid: v.Valid, Float64: v.Float64})
		default:
			nvArgs = append(nvArgs, arg)
		}
	}

	// 带有慢查询检测
	startTime := time.Now()
	result, err := db.Exec(query, nvArgs...)
	duration := time.Since(startTime)

	if duration > defaultSlowThreshold {
		logx.WithDuration(duration).Slowf("[SQL] 慢执行(%v) - %+v", duration, stmt)
	} else {
		logx.WithDuration(duration).Infof("[SQL] 执行: %+v", stmt)
	}

	if err != nil {
		logSqlError(stmt, err)
	}

	return result, err
}
