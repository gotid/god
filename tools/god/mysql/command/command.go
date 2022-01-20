package command

import (
	"errors"
	"strings"

	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/tools/god/mysql/gen"
	"github.com/gotid/god/tools/god/mysql/model"
	"github.com/gotid/god/tools/god/util"
	"github.com/urfave/cli"
)

const (
	flagDSN   = "dsn"
	flagTable = "table"
	flagDir   = "dir"
	flagCache = "cache"
)

func GenCodeFromDSN(ctx *cli.Context) error {
	dsn := strings.TrimSpace(ctx.String(flagDSN))
	dir := strings.TrimSpace(ctx.String(flagDir))
	cache := ctx.Bool(flagCache)
	table := strings.TrimSpace(ctx.String(flagTable))

	logx.Disable()
	log := util.NewConsole(true)

	if len(dsn) == 0 {
		log.Error("MySQL连接地址未提供")
		return nil
	}
	if len(table) == 0 {
		log.Error("表名未提供")
		return nil
	}

	tables := collection.NewSet()
	for _, table = range strings.Split(table, ",") {
		table = strings.TrimSpace(table)
		if len(table) == 0 {
			continue
		}
		tables.AddStr(table)
	}

	// 获取数据库名称
	path := strings.Split(dsn, "?")[0]
	parts := strings.Split(path, "/")
	database := strings.TrimSpace(parts[len(parts)-1])
	if !strings.Contains(path, "/") || database == "" {
		log.Error("数据库连接字符串：未提供数据库名称")
		return errors.New("数据库连接字符串：未提供数据库名称")
	}

	conn := sqlx.NewMySQL(dsn)
	m := model.NewModel(conn)
	ddlList, err := m.ShowDDL(tables.KeysStr()...)
	if err != nil {
		log.Error("", err)
		return nil
	}

	// fmt.Println(strings.Join(ddlList, "\n"), dir, cache)
	generator := gen.NewModelGenerator(ddlList, dir, gen.WithConsoleOption(log), gen.WithDatabaseOption(database))
	err = generator.Start(cache)
	if err != nil {
		log.Error("", err)
	}

	return nil
}
