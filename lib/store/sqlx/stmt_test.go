package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var errMockedPlaceholder = errors.New("placeholder")

func TestStmt_exec(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		args         []any
		delay        bool
		hasError     bool
		err          error
		lastInsertId int64
		rowsAffected int64
	}{
		{
			name:         "normal",
			query:        "select user from users where id=?",
			args:         []any{1},
			lastInsertId: 1,
			rowsAffected: 2,
		},
		{
			name:     "exec error",
			query:    "select user from users where id=?",
			args:     []any{1},
			hasError: true,
			err:      errors.New("exec"),
		},
		{
			name:     "exec more args error",
			query:    "select user from users where id=? and name=?",
			args:     []any{1},
			hasError: true,
			err:      errors.New("exec"),
		},
		{
			name:         "慢查询",
			query:        "select user from users where id=?",
			args:         []any{1},
			delay:        true,
			lastInsertId: 1,
			rowsAffected: 2,
		},
	}

	for _, test := range tests {
		test := test
		fns := []func(args ...any) (sql.Result, error){
			func(args ...any) (sql.Result, error) {
				return exec(context.Background(), &mockedSessionConn{
					lastInsertId: test.lastInsertId,
					rowsAffected: test.rowsAffected,
					err:          test.err,
					delay:        test.delay,
				}, test.query, args...)
			},
			func(args ...any) (sql.Result, error) {
				return execStmt(context.Background(), &mockedStmtConn{
					lastInsertId: test.lastInsertId,
					rowsAffected: test.rowsAffected,
					err:          test.err,
					delay:        test.delay,
				}, test.query, args...)
			},
		}

		for _, fn := range fns {
			fn := fn
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				res, err := fn(test.args...)
				if test.hasError {
					assert.NotNil(t, err)
					return
				}

				assert.Nil(t, err)
				lastInsertId, err := res.LastInsertId()
				assert.Nil(t, err)
				assert.Equal(t, test.lastInsertId, lastInsertId)
				rowsAffected, err := res.RowsAffected()
				assert.Nil(t, err)
				assert.Equal(t, test.rowsAffected, rowsAffected)
			})
		}
	}
}

func TestStmt_query(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		args     []any
		delay    bool
		hasError bool
		err      error
	}{
		{
			name:  "normal",
			query: "select user from users where id=?",
			args:  []any{1},
		},
		{
			name:     "query error",
			query:    "select user from users where id=?",
			args:     []any{1},
			hasError: true,
			err:      errors.New("exec"),
		},
		{
			name:     "query more args error",
			query:    "select user from users where id=? and name=?",
			args:     []any{1},
			hasError: true,
			err:      errors.New("exec"),
		},
		{
			name:  "慢查询",
			query: "select user from users where id=?",
			args:  []any{1},
			delay: true,
		},
	}

	for _, test := range tests {
		test := test
		fns := []func(args ...any) error{
			func(args ...any) error {
				return query(context.Background(), &mockedSessionConn{
					err:   test.err,
					delay: test.delay,
				}, func(rows *sql.Rows) error {
					return nil
				}, test.query, args...)
			},
			func(args ...any) error {
				return queryStmt(context.Background(), &mockedStmtConn{
					err:   test.err,
					delay: test.delay,
				}, func(rows *sql.Rows) error {
					return nil
				}, test.query, args...)
			},
		}

		for _, fn := range fns {
			fn := fn
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				err := fn(test.args...)
				if test.hasError {
					assert.NotNil(t, err)
					return
				}

				assert.NotNil(t, err)
			})
		}
	}
}

func TestSetSlowThreshold(t *testing.T) {
	assert.Equal(t, defaultSlowThreshold, slowThreshold.Load())
	SetSlowThreshold(time.Second)
	assert.Equal(t, time.Second, slowThreshold.Load())
}

func TestDisableLog(t *testing.T) {
	assert.True(t, logSQL.True())
	assert.True(t, logSlowSQL.True())
	defer func() {
		logSQL.Set(true)
		logSlowSQL.Set(true)
	}()

	DisableLog()
	assert.False(t, logSQL.True())
	assert.False(t, logSlowSQL.True())
}

func TestDisableStmtLog(t *testing.T) {
	assert.True(t, logSQL.True())
	assert.True(t, logSlowSQL.True())
	defer func() {
		logSQL.Set(true)
		logSlowSQL.Set(true)
	}()

	DisableStmtLog()
	assert.False(t, logSQL.True())
	assert.True(t, logSlowSQL.True())
}

func TestNilGuard(t *testing.T) {
	assert.True(t, logSQL.True())
	assert.True(t, logSlowSQL.True())
	defer func() {
		logSQL.Set(true)
		logSlowSQL.Set(true)
	}()

	DisableLog()
	guard := newGuard("any")
	assert.Nil(t, guard.start("foo", "bar"))
	guard.finish(context.Background(), nil)
	assert.Equal(t, nilGuard{}, guard)
}

type mockedSessionConn struct {
	lastInsertId int64
	rowsAffected int64
	err          error
	delay        bool
}

func (m *mockedSessionConn) Exec(query string, args ...any) (sql.Result, error) {
	return m.ExecContext(context.Background(), query, args...)
}

func (m *mockedSessionConn) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	if m.delay {
		time.Sleep(defaultSlowThreshold + time.Millisecond)
	}
	return mockedResult{
		lastInsertId: m.lastInsertId,
		rowsAffected: m.rowsAffected,
	}, m.err
}

func (m *mockedSessionConn) Query(query string, args ...any) (*sql.Rows, error) {
	return m.QueryContext(context.Background(), query, args...)
}

func (m *mockedSessionConn) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) {
	if m.delay {
		time.Sleep(defaultSlowThreshold + time.Millisecond)
	}

	err := errMockedPlaceholder
	if m.err != nil {
		err = m.err
	}
	return new(sql.Rows), err
}

type mockedStmtConn struct {
	lastInsertId int64
	rowsAffected int64
	err          error
	delay        bool
}

func (m *mockedStmtConn) Exec(args ...any) (sql.Result, error) {
	return m.ExecContext(context.Background(), args...)
}

func (m *mockedStmtConn) ExecContext(_ context.Context, _ ...any) (sql.Result, error) {
	if m.delay {
		time.Sleep(defaultSlowThreshold + time.Millisecond)
	}
	return mockedResult{
		lastInsertId: m.lastInsertId,
		rowsAffected: m.rowsAffected,
	}, m.err
}

func (m *mockedStmtConn) Query(args ...any) (*sql.Rows, error) {
	return m.QueryContext(context.Background(), args...)
}

func (m *mockedStmtConn) QueryContext(_ context.Context, _ ...any) (*sql.Rows, error) {
	if m.delay {
		time.Sleep(defaultSlowThreshold + time.Millisecond)
	}

	err := errMockedPlaceholder
	if m.err != nil {
		err = m.err
	}
	return new(sql.Rows), err
}

type mockedResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (m mockedResult) LastInsertId() (int64, error) {
	return m.lastInsertId, nil
}

func (m mockedResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}
