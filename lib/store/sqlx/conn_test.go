package sqlx

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gotid/god/lib/logx"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

const mockedDatasource = "sqlmock"

func init() {
	logx.Disable()
}

func TestConn(t *testing.T) {
	mock, err := buildConn()
	assert.Nil(t, err)

	mock.ExpectExec("any")
	mock.ExpectQuery("any").WillReturnRows(sqlmock.NewRows([]string{"foo"}))

	conn := NewMySQL(mockedDatasource)
	assert.Nil(t, err)
	_, err = conn.Exec("any", "value")
	assert.NotNil(t, err)

	badConn := NewMySQL("bad-sql")
	_, err = badConn.Exec("any", "value")
	assert.NotNil(t, err)
	_, err = badConn.Prepare("any")
	assert.NotNil(t, err)

	db, err := conn.RawDB()
	rawConn := NewConnFromDB(db, withMySQLAcceptable())
	_, err = rawConn.Prepare("any")
	assert.NotNil(t, err)

	var val string
	assert.NotNil(t, conn.QueryRow(&val, "any"))
	assert.NotNil(t, badConn.QueryRow(&val, "any"))
	assert.NotNil(t, conn.QueryRowPartial(&val, "any"))
	assert.NotNil(t, badConn.QueryRowPartial(&val, "any"))
	assert.NotNil(t, conn.QueryRows(&val, "any"))
	assert.NotNil(t, badConn.QueryRows(&val, "any"))
	assert.NotNil(t, conn.QueryRowsPartial(&val, "any"))
	assert.NotNil(t, badConn.QueryRowsPartial(&val, "any"))
	assert.NotNil(t, conn.Transact(func(session Session) error {
		return nil
	}))
	assert.NotNil(t, badConn.Transact(func(session Session) error {
		return nil
	}))
}

func buildConn() (mock sqlmock.Sqlmock, err error) {
	_, err = connManager.Get(mockedDatasource, func() (io.Closer, error) {
		var db *sql.DB
		var err error
		db, mock, err = sqlmock.New()
		return &pingedDB{
			DB: db,
		}, err
	})

	return
}
