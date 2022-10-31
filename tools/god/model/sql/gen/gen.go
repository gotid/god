package gen

import (
	"bytes"
	"fmt"
	"github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/model/sql/model"
	"github.com/gotid/god/tools/god/model/sql/parser"
	"github.com/gotid/god/tools/god/model/sql/template"
	util2 "github.com/gotid/god/tools/god/model/sql/util"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/console"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"os"
	"path/filepath"
	"strings"
)

const pwd = "."

type (
	DefaultGenerator struct {
		console.Console
		dir           string
		pkg           string
		cfg           *config.Config
		isPostgreSQL  bool
		ignoreColumns []string
	}

	// Option 自定义生成器参数。
	Option func(*DefaultGenerator)

	// Table 描述一张 mysql 表。
	Table struct {
		parser.Table
		PrimaryCacheKey        Key
		UniqueCacheKey         []Key
		ContainsUniqueCacheKey bool
		ignoreColumns          []string
	}

	code struct {
		importsCode string
		varsCode    string
		typesCode   string
		newCode     string
		insertCode  string
		findCode    []string
		updateCode  string
		deleteCode  string
		cacheExtra  string
		tableName   string
	}

	codeTuple struct {
		modelCode       string
		modelCustomCode string
	}
)

// NewDefaultGenerator 返回一个默认生成器实例 defaultGenerator。
func NewDefaultGenerator(dir string, cfg *config.Config, opt ...Option) (*DefaultGenerator, error) {
	if dir == "" {
		dir = pwd
	}
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	dir = dirAbs
	pkg := util.SafeString(filepath.Base(dirAbs))
	err = pathx.MkdirIfNotExist(dir)
	if err != nil {
		return nil, err
	}

	generator := &DefaultGenerator{dir: dir, cfg: cfg, pkg: pkg}
	var opts []Option
	opts = append(opts, newDefaultOption())
	opts = append(opts, opt...)
	for _, fn := range opts {
		fn(generator)
	}

	return generator, err

}

// WithConsoleOption 自定义控制台。
func WithConsoleOption(c console.Console) Option {
	return func(generator *DefaultGenerator) {
		generator.Console = c
	}
}

// WithIgnoreColumns 自定义插入或更新时的要忽略的列。
func WithIgnoreColumns(ignoreColumns []string) Option {
	return func(generator *DefaultGenerator) {
		generator.ignoreColumns = ignoreColumns
	}
}

// WithPostgreSQL 作为 postgresql 生成器。
func WithPostgreSQL() Option {
	return func(generator *DefaultGenerator) {
		generator.isPostgreSQL = true
	}
}

// StartFromDDL 从 DDL 脚本文件生成模型。
func (g *DefaultGenerator) StartFromDDL(filename string, withCache, strict bool, database string) error {
	models, err := g.genFromDDL(filename, withCache, strict, database)
	if err != nil {
		return err
	}

	return g.createFile(models)
}

// StartFromInformationSchema 从给定的表中生成模型。
func (g *DefaultGenerator) StartFromInformationSchema(tables map[string]*model.Table, withCache, strict bool) error {
	m := make(map[string]*codeTuple)
	for _, each := range tables {
		table, err := parser.ConvertDataType(each, strict)
		if err != nil {
			return err
		}

		code, err := g.genModel(*table, withCache)
		if err != nil {
			return err
		}

		customCode, err := g.genModelCustom(*table, withCache)
		if err != nil {
			return err
		}

		m[table.Name.Source()] = &codeTuple{
			modelCode:       code,
			modelCustomCode: customCode,
		}
	}

	return g.createFile(m)
}

func (g *DefaultGenerator) genFromDDL(filename string, withCache, strict bool, database string) (map[string]*codeTuple, error) {
	m := make(map[string]*codeTuple)
	tables, err := parser.Parse(filename, database, strict)
	if err != nil {
		return nil, err
	}

	for _, table := range tables {
		code, err := g.genModel(*table, withCache)
		if err != nil {
			return nil, err
		}

		customCode, err := g.genModelCustom(*table, withCache)
		if err != nil {
			return nil, err
		}

		m[table.Name.Source()] = &codeTuple{
			modelCode:       code,
			modelCustomCode: customCode,
		}
	}

	return m, nil
}

func (g *DefaultGenerator) genModel(t parser.Table, withCache bool) (string, error) {
	if len(t.PrimaryKey.Name.Source()) == 0 {
		return "", fmt.Errorf("表 %s：缺失主键", t.Name.Source())
	}

	primaryKey, uniqueKey := genCacheKeys(t)

	table := Table{
		Table:                  t,
		PrimaryCacheKey:        primaryKey,
		UniqueCacheKey:         uniqueKey,
		ContainsUniqueCacheKey: len(uniqueKey) > 0,
		ignoreColumns:          g.ignoreColumns,
	}

	importsCode, err := genImports(table, withCache, t.ContainsTime())
	if err != nil {
		return "", err
	}

	varsCode, err := genVars(table, withCache, g.isPostgreSQL)
	if err != nil {
		return "", err
	}

	insertCode, insertCodeMethod, err := genInsert(table, withCache, g.isPostgreSQL)
	if err != nil {
		return "", err
	}

	findCode := make([]string, 0)
	findOneCode, findOneCodeMethod, err := genFindOne(table, withCache, g.isPostgreSQL)
	if err != nil {
		return "", err
	}

	ret, err := genFindOneByField(table, withCache, g.isPostgreSQL)
	if err != nil {
		return "", err
	}

	findCode = append(findCode, findOneCode, ret.findOneMethod)

	updateCode, updateCodeMethod, err := genUpdate(table, withCache, g.isPostgreSQL)
	if err != nil {
		return "", err
	}

	deleteCode, deleteCodeMethod, err := genDelete(table, withCache, g.isPostgreSQL)
	if err != nil {
		return "", err
	}

	var list []string
	list = append(list, insertCodeMethod, findOneCodeMethod, ret.findOneInterfaceMethod,
		updateCodeMethod, deleteCodeMethod)
	typesCode, err := genTypes(table, strings.Join(util2.TrimStringSlice(list), pathx.NL), withCache)
	if err != nil {
		return "", err
	}

	newCode, err := genNew(table, withCache, g.isPostgreSQL)
	if err != nil {
		return "", err
	}

	tableName, err := genTableName(table)
	if err != nil {
		return "", err
	}

	code := &code{
		importsCode: importsCode,
		varsCode:    varsCode,
		typesCode:   typesCode,
		newCode:     newCode,
		insertCode:  insertCode,
		findCode:    findCode,
		updateCode:  updateCode,
		deleteCode:  deleteCode,
		cacheExtra:  ret.cacheExtra,
		tableName:   tableName,
	}

	output, err := g.executeModel(table, code)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func (g *DefaultGenerator) genModelCustom(table parser.Table, withCache bool) (string, error) {
	text, err := pathx.LoadTemplate(category, modelCustomTemplateFile, template.ModelCustom)
	if err != nil {
		return "", err
	}

	output, err := util.With("model-custom").Parse(text).GoFmt(true).Execute(map[string]interface{}{
		"pkg":                   g.pkg,
		"withCache":             withCache,
		"upperStartCamelObject": table.Name.ToCamel(),
		"lowerStartCamelObject": stringx.From(table.Name.ToCamel()).UnTitle(),
	})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func (g *DefaultGenerator) executeModel(table Table, code *code) (*bytes.Buffer, error) {
	text, err := pathx.LoadTemplate(category, modelGenTemplateFile, template.ModelGen)
	if err != nil {
		return nil, err
	}

	output, err := util.With("model").Parse(text).GoFmt(true).Execute(map[string]interface{}{
		"pkg":         g.pkg,
		"imports":     code.importsCode,
		"vars":        code.varsCode,
		"types":       code.typesCode,
		"new":         code.newCode,
		"insert":      code.insertCode,
		"find":        strings.Join(code.findCode, "\n"),
		"update":      code.updateCode,
		"delete":      code.deleteCode,
		"extraMethod": code.cacheExtra,
		"tableName":   code.tableName,
		"data":        table,
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (g *DefaultGenerator) createFile(models map[string]*codeTuple) error {
	dirAbs, err := filepath.Abs(g.dir)
	if err != nil {
		return err
	}

	g.dir = dirAbs
	g.pkg = util.SafeString(filepath.Base(dirAbs))
	err = pathx.MkdirIfNotExist(dirAbs)
	if err != nil {
		return err
	}

	// 生成每张表的模型文件
	for tableName, codes := range models {
		tn := stringx.From(tableName)
		modelFilename, err := format.FileNamingFormat(g.cfg.NamingFormat, fmt.Sprintf("%s_model", tn.Source()))
		if err != nil {
			return err
		}

		name := util.SafeString(modelFilename) + "_gen.go"
		filename := filepath.Join(dirAbs, name)
		err = os.WriteFile(filename, []byte(codes.modelCode), os.ModePerm)
		if err != nil {
			return err
		}

		name = util.SafeString(modelFilename) + ".go"
		filename = filepath.Join(dirAbs, name)
		if pathx.FileExists(filename) {
			g.Warning("%s 已存在，忽略。", name)
		}
		err = os.WriteFile(filename, []byte(codes.modelCustomCode), os.ModePerm)
		if err != nil {
			return err
		}
	}

	// 生成变量文件
	varFilename, err := format.FileNamingFormat(g.cfg.NamingFormat, "vars")
	if err != nil {
		return err
	}
	filename := filepath.Join(dirAbs, varFilename+".go")
	text, err := pathx.LoadTemplate(category, errTemplateFile, template.Error)
	if err != nil {
		return err
	}
	err = util.With("vars").Parse(text).SaveTo(map[string]interface{}{
		"pkg": g.pkg,
	}, filename, false)
	if err != nil {
		return err
	}

	g.Success("完成。")
	return nil
}

func (t Table) isIgnoreColumns(columnName string) bool {
	for _, v := range t.ignoreColumns {
		if v == columnName {
			return true
		}
	}
	return false
}

func newDefaultOption() Option {
	return func(generator *DefaultGenerator) {
		generator.Console = console.NewColorConsole()
	}
}

func wrapWithRawString(s string, postgreSql bool) string {
	if postgreSql || s == "`" {
		return s
	}

	if !strings.HasPrefix(s, "`") {
		s = "`" + s
	}

	if !strings.HasSuffix(s, "`") {
		s = s + "`"
	} else if len(s) == 1 {
		s = s + "`"
	}

	return s
}
