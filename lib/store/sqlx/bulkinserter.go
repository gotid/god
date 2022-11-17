package sqlx

import (
	"database/sql"
	"fmt"
	"github.com/gotid/god/lib/executors"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/stringx"
	"strings"
	"time"
)

const (
	valuesKeyword = "values"
	flushInterval = time.Second
	maxBulkRows   = 1000
)

var emptyBulkStmt bulkStmt

type (
	// BulkInserter 用于批量插入记录。
	// 暂不支持 pg 和 oracle，因为它俩使用 `$` 和 `:` 进行格式化。
	BulkInserter struct {
		executor *executors.PeriodicalExecutor
		inserter *dbInserter
		stmt     bulkStmt
	}

	// ResultHandler 是一个 sql.Result 结果处理函数。
	ResultHandler func(sql.Result, error)

	bulkStmt struct {
		prefix      string
		valueFormat string
		suffix      string
	}
)

// NewBulkInserter 返回一个批量插入器 BulkInserter。
func NewBulkInserter(conn Conn, stmt string) (*BulkInserter, error) {
	bkStmt, err := parseInsertStmt(stmt)
	if err != nil {
		return nil, err
	}

	inserter := &dbInserter{
		conn: conn,
		stmt: bkStmt,
	}

	return &BulkInserter{
		executor: executors.NewPeriodicalExecutor(flushInterval, inserter),
		inserter: inserter,
		stmt:     bkStmt,
	}, nil
}

// Flush 流转所有挂起的任务。
func (bi *BulkInserter) Flush() {
	bi.executor.Flush()
}

// Insert 插入给定的参数。
func (bi *BulkInserter) Insert(args ...any) error {
	value, err := format(bi.stmt.valueFormat, args...)
	if err != nil {
		return err
	}

	bi.executor.Add(value)

	return nil
}

// SetResultHandler 设置 sql.Result 结果处理器。
func (bi *BulkInserter) SetResultHandler(handler ResultHandler) {
	bi.executor.Sync(func() {
		bi.inserter.resultHandler = handler
	})
}

// UpdateOrDelete 在流转挂起任务后，更新或删除查询。
func (bi *BulkInserter) UpdateOrDelete(fn func()) {
	bi.executor.Flush()
	fn()
}

// UpdateStmt 更新插入语句。
func (bi *BulkInserter) UpdateStmt(stmt string) error {
	bkStmt, err := parseInsertStmt(stmt)
	if err != nil {
		return err
	}

	bi.executor.Flush()
	bi.executor.Sync(func() {
		bi.inserter.stmt = bkStmt
	})

	return nil
}

// db 插入器的任务容器
type dbInserter struct {
	conn          Conn
	stmt          bulkStmt
	values        []string
	resultHandler ResultHandler
}

func (in *dbInserter) AddTask(task any) bool {
	in.values = append(in.values, task.(string))
	return len(in.values) >= maxBulkRows
}

func (in *dbInserter) Execute(tasks any) {
	values := tasks.([]string)
	if len(values) == 0 {
		return
	}

	stmtWithoutValues := in.stmt.prefix
	valueStr := strings.Join(values, ", ")
	stmt := strings.Join([]string{stmtWithoutValues, valueStr}, " ")
	if len(in.stmt.suffix) > 0 {
		stmt = strings.Join([]string{stmt, in.stmt.suffix}, " ")
	}
	result, err := in.conn.Exec(stmt)
	if in.resultHandler != nil {
		in.resultHandler(result, err)
	} else if err != nil {
		logx.Errorf("SQL：%s，执行错误：%s", stmt, err)
	}
}

func (in *dbInserter) RemoveAll() any {
	values := in.values
	in.values = nil
	return values
}

func parseInsertStmt(stmt string) (bulkStmt, error) {
	lower := strings.ToLower(stmt)
	pos := strings.Index(lower, valuesKeyword)
	if pos <= 0 {
		return emptyBulkStmt, fmt.Errorf("错误的 SQL 插入语句：%q", stmt)
	}

	var columns int
	right := strings.LastIndexByte(lower[:pos], ')')
	if right > 0 {
		left := strings.LastIndexByte(lower[:right], '(')
		if left > 0 {
			values := lower[left+1 : right]
			values = stringx.Filter(values, func(r rune) bool {
				return r == ' ' || r == '\t' || r == '\r' || r == '\n'
			})
			fields := strings.FieldsFunc(values, func(r rune) bool {
				return r == ','
			})
			columns = len(fields)
		}
	}

	var variables int
	var valueFormat string
	var suffix string
	left := strings.IndexByte(lower[pos:], '(')
	if left > 0 {
		right = strings.IndexByte(lower[pos+left:], ')')
		if right > 0 {
			values := lower[pos+left : pos+left+right]
			for _, x := range values {
				if x == '?' {
					variables++
				}
			}
			valueFormat = stmt[pos+left : pos+left+right+1]
			suffix = strings.TrimSpace(stmt[pos+left+right+1:])
		}
	}

	if variables == 0 {
		return emptyBulkStmt, fmt.Errorf("SQL 插入语句没有变量：%q", stmt)
	}
	if columns > 0 && columns != variables {
		return emptyBulkStmt, fmt.Errorf("字段和变量的个数不一致：%q", stmt)
	}

	return bulkStmt{
		prefix:      stmt[:pos+len(valuesKeyword)],
		valueFormat: valueFormat,
		suffix:      suffix,
	}, nil
}
