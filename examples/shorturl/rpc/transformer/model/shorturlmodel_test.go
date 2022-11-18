package model

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gotid/god/lib/jsonx"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/store/sqlc"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultShorturlModel_Insert(t *testing.T) {
	ast := assert.New(t)

	// 构建模型，模仿数据库和redis
	model, mock, _ := newMockShorturlModel(t)

	// 模拟测试情况
	mock.ExpectExec(fmt.Sprintf("insert into %s", model.table)).
		WithArgs("123", "123").
		WillReturnError(errors.New("执行错误"))
	mock.ExpectExec(fmt.Sprintf("insert into %s", model.table)).
		WithArgs("345", "345").
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()

	_, err := model.Insert(ctx, &Shorturl{
		Shorten: "123",
		Url:     "123",
	})
	ast.NotNil(err)

	_, err = model.Insert(ctx, &Shorturl{
		Shorten: "345",
		Url:     "345",
	})
	ast.Nil(err)
}

func TestDefaultShorturlModel_Update(t *testing.T) {
	ast := assert.New(t)

	// 构建模型，模仿数据库和redis
	model, mock, _ := newMockShorturlModel(t)

	ctx := context.Background()

	// 模拟更新失败情况
	mock.ExpectExec(fmt.Sprintf("update %s", model.table)).
		WillReturnError(errors.New("执行错误"))
	err := model.Update(ctx, &Shorturl{
		Shorten: "123",
		Url:     "123",
	})
	ast.NotNil(err)

	// 模拟更新成功情况
	mock.ExpectExec(fmt.Sprintf("update %s", model.table)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err = model.Update(ctx, &Shorturl{
		Shorten: "123",
		Url:     "123",
	})
	ast.Nil(err)
}

func TestDefaultShorturlModel_Delete(t *testing.T) {
	ast := assert.New(t)

	// build model, mock db and mock redis
	model, mock, _ := newMockShorturlModel(t)

	ctx := context.Background()

	// mock test fail case
	mock.ExpectExec(fmt.Sprintf("delete from  %s", model.table)).
		WillReturnError(errors.New("exec error"))
	err := model.Delete(ctx, "123")
	ast.NotNil(err)

	// mock test success case
	mock.ExpectExec(fmt.Sprintf("delete from  %s", model.table)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err = model.Delete(ctx, "123")
	ast.Nil(err)
}

func TestDefaultShorturlModel_FindOne(t *testing.T) {
	ast := assert.New(t)

	// build model, mock db and mock redis
	model, mock, rds := newMockShorturlModel(t)

	ctx := context.Background()

	// mock db query error
	mock.ExpectQuery(fmt.Sprintf("select (.+) from %s", model.table)).
		WillReturnError(errors.New("query error"))

	_, err := model.FindOne(ctx, "123")
	ast.NotNil(err)

	// mock db query success
	rows := sqlmock.NewRows(
		[]string{"shorten", "url"},
	).AddRow([]driver.Value{"111", "222"}...)

	mock.ExpectQuery(fmt.Sprintf("select (.+) from %s", model.table)).
		WillReturnRows(rows)

	ret, err := model.FindOne(ctx, "111")
	ast.Nil(err)
	ast.Equal(ret, &Shorturl{
		Shorten: "111",
		Url:     "222",
	})

	// mock cache data
	su := &Shorturl{
		Shorten: "123",
		Url:     "234",
	}
	data, _ := jsonx.Marshal(su)
	rds.Set(fmt.Sprintf("%s%v", cacheShorturlShortenPrefix, su.Shorten), string(data))

	ret, err = model.FindOne(ctx, su.Shorten)
	ast.Nil(err)
	ast.Equal(ret, su)
}

func newMockShorturlModel(t *testing.T) (*defaultShorturlModel, sqlmock.Sqlmock, *miniredis.Miniredis) {
	db, mockDb, err := sqlmock.New()
	if err != nil {
		t.Fatalf("打开模拟数据库连接时出现错误 '%s'", err)
	}

	rds := miniredis.RunT(t)
	return &defaultShorturlModel{
		CachedConn: sqlc.NewNodeConn(sqlx.NewConnFromDB(db), redis.New(rds.Addr())),
		table:      "`api`",
	}, mockDb, rds
}
