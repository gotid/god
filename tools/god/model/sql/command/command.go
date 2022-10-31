package command

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/tools/god/model/sql/gen"
	"github.com/gotid/god/tools/god/model/sql/model"
	"path/filepath"
	"strings"

	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/model/sql/util"
	file "github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/console"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/spf13/cobra"
)

var (
	// VarStringSrc sql 源文件
	VarStringSrc string
	// VarStringDir sql 输出目录
	VarStringDir string
	// VarBoolCache 是否启用缓存
	VarBoolCache bool
	// VarBoolIdea 是否在 idea IDE 中
	VarBoolIdea bool
	// VarStringURL 数据源链接地址
	VarStringURL string
	// VarStringSliceTable 数据表
	VarStringSliceTable []string
	// VarStringStyle 文件命名风格
	VarStringStyle string
	// VarStringDatabase 数据库
	VarStringDatabase string
	// VarStringSchema postgresql 的概要
	VarStringSchema string
	// VarStringHome god 代码生成器的主目录
	VarStringHome string
	// VarStringRemote git 仓库地址
	VarStringRemote string
	// VarStringBranch 仓库分支
	VarStringBranch string
	// VarBoolStrict 是否启用严格模式
	VarBoolStrict bool
	// VarStringSliceIgnoreColumns 忽略的列
	VarStringSliceIgnoreColumns []string
)

var errNotMatched = errors.New("未找到匹配的SQL脚本文件")

// MySqlDDL 从 ddl 脚本生成模型。
func MySqlDDL(_ *cobra.Command, _ []string) error {
	src := VarStringSrc
	dir := VarStringDir
	cache := VarBoolCache
	idea := VarBoolIdea
	style := VarStringStyle
	database := VarStringDatabase
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	if len(remote) > 0 {
		repo, _ := file.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGodHome(home)
	}
	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	arg := ddlArg{
		src:           src,
		dir:           dir,
		cfg:           cfg,
		cache:         cache,
		idea:          idea,
		database:      database,
		strict:        VarBoolStrict,
		ignoreColumns: mergeColumns(VarStringSliceIgnoreColumns),
	}

	return fromDDL(arg)
}

func fromDDL(arg ddlArg) error {
	log := console.NewConsole(arg.idea)
	src := strings.TrimSpace(arg.src)
	if len(src) == 0 {
		return errors.New("未指定SQL脚本路径或通配符样式")
	}

	files, err := util.MatchFiles(src)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return errNotMatched
	}

	generator, err := gen.NewDefaultGenerator(arg.dir, arg.cfg,
		gen.WithConsoleOption(log), gen.WithIgnoreColumns(arg.ignoreColumns))
	if err != nil {
		return err
	}

	for _, filename := range files {
		err = generator.StartFromDDL(filename, arg.cache, arg.strict, arg.database)
		if err != nil {
			return err
		}
	}

	return nil
}

func mergeColumns(columns []string) []string {
	set := collection.NewSet()
	for _, v := range columns {
		fields := strings.FieldsFunc(v, func(r rune) bool {
			return r == ','
		})
		set.AddStr(fields...)
	}

	return set.KeysStr()
}

func MySqlDSN(_ *cobra.Command, _ []string) error {
	url := strings.TrimSpace(VarStringURL)
	dir := strings.TrimSpace(VarStringDir)
	cache := VarBoolCache
	idea := VarBoolIdea
	style := VarStringStyle
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	if len(remote) > 0 {
		repo, _ := file.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}
	if len(home) > 0 {
		pathx.RegisterGodHome(home)
	}

	tableValue := VarStringSliceTable
	patterns := parseTableList(tableValue)
	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	arg := dataSourceArg{
		url:           url,
		dir:           dir,
		tablePat:      patterns,
		cfg:           cfg,
		cache:         cache,
		idea:          idea,
		strict:        VarBoolStrict,
		ignoreColumns: mergeColumns(VarStringSliceIgnoreColumns),
	}

	return fromMysqlDataSource(arg)
}

func fromMysqlDataSource(arg dataSourceArg) error {
	log := console.NewConsole(arg.idea)
	if len(arg.url) == 0 {
		log.Error("%v", "未提供 MySQL DSN 数据源")
		return nil
	}

	if len(arg.tablePat) == 0 {
		log.Error("%v", "未指定表或表的通配符样式")
		return nil
	}

	dsn, err := mysql.ParseDSN(arg.url)
	if err != nil {
		return err
	}

	logx.Disable()
	databaseSource := strings.TrimSuffix(arg.url, "/"+dsn.DBName) + "/information_schema"
	db := sqlx.NewMySQL(databaseSource)
	im := model.NewInformationSchemaModel(db)

	tables, err := im.GetAllTables(dsn.DBName)
	if err != nil {
		return err
	}

	matchTables := make(map[string]*model.Table)
	for _, item := range tables {
		if !arg.tablePat.Match(item) {
			continue
		}

		columnData, err := im.FindColumns(dsn.DBName, item)
		if err != nil {
			return err
		}

		table, err := columnData.Convert()
		if err != nil {
			return err
		}

		matchTables[item] = table
	}

	if len(matchTables) == 0 {
		return errors.New("没有匹配的表")
	}

	generator, err := gen.NewDefaultGenerator(arg.dir, arg.cfg,
		gen.WithConsoleOption(log), gen.WithIgnoreColumns(arg.ignoreColumns))
	if err != nil {
		return err
	}

	return generator.StartFromInformationSchema(matchTables, arg.cache, arg.strict)
}

type dataSourceArg struct {
	url, dir      string
	tablePat      pattern
	cfg           *config.Config
	cache, idea   bool
	strict        bool
	ignoreColumns []string
}

type pattern map[string]struct{}

func (p pattern) Match(s string) bool {
	for v := range p {
		match, err := filepath.Match(v, s)
		if err != nil {
			console.Error("%+v", err)
			continue
		}
		if match {
			return true
		}
	}
	return false
}

func (p pattern) list() []string {
	var ret []string
	for v := range p {
		ret = append(ret, v)
	}
	return ret
}

func parseTableList(tableValue []string) pattern {
	tablePattern := make(pattern)
	for _, v := range tableValue {
		fields := strings.FieldsFunc(v, func(r rune) bool {
			return r == ','
		})
		for _, f := range fields {
			tablePattern[f] = struct{}{}
		}
	}
	return tablePattern
}

type ddlArg struct {
	src, dir      string
	cfg           *config.Config
	cache, idea   bool
	database      string
	strict        bool
	ignoreColumns []string
}
