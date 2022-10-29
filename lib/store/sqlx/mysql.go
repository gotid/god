package sqlx

import "github.com/go-sql-driver/mysql"

const (
	mysqlDriverName           = "mysql"
	duplicateEntryCode uint16 = 1062
)

// NewMySQL 返回一个 MySQL 连接。
// dataSourceName 为 mysql/sqlite/sqlmock/clickhouse 等。
func NewMySQL(dataSourceName string, opts ...Option) Conn {
	opts = append(opts, withMySQLAcceptable())
	return NewConn(mysqlDriverName, dataSourceName, opts...)
}

func withMySQLAcceptable() Option {
	return func(conn *commonConn) {
		conn.accept = mysqlAcceptable
	}
}

func mysqlAcceptable(err error) bool {
	if err == nil {
		return true
	}

	myErr, ok := err.(*mysql.MySQLError)
	if !ok {
		return false
	}

	switch myErr.Number {
	case duplicateEntryCode:
		return true
	default:
		return false
	}
}
