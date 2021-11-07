package gen

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"git.zc0901.com/go/god/lib/fs"
	"git.zc0901.com/go/god/lib/stringx"
	"git.zc0901.com/go/god/tools/god/mysql/parser"
	"git.zc0901.com/go/god/tools/god/mysql/tpl"
	"git.zc0901.com/go/god/tools/god/util"
)

const (
	pwd = "."
)

type (
	ModelGenerator struct {
		database string
		ddlList  []string
		dir      string
		util.Console
	}

	Option func(gen *ModelGenerator)

	Table struct {
		parser.Table
		CacheKeys map[string]Key
	}
)

func NewModelGenerator(ddlList []string, dir string, opts ...Option) *ModelGenerator {
	if dir == "" {
		dir = pwd
	}
	generator := &ModelGenerator{ddlList: ddlList, dir: dir}
	var optionList []Option
	optionList = append(optionList, newDefaultOption())
	optionList = append(optionList, opts...)
	for _, fn := range optionList {
		fn(generator)
	}
	return generator
}

func newDefaultOption() Option {
	return func(gen *ModelGenerator) {
		gen.Console = util.NewColorConsole()
	}
}

func WithConsoleOption(c util.Console) Option {
	return func(gen *ModelGenerator) {
		gen.Console = c
	}
}

func WithDatabaseOption(database string) Option {
	return func(gen *ModelGenerator) {
		gen.database = database
	}
}

func (g *ModelGenerator) Start(withCache bool) error {
	dir, err := filepath.Abs(g.dir)
	if err != nil {
		return err
	}
	if err = fs.MkdirIfNotExist(dir); err != nil {
		return err
	}

	modelList, err := g.genFromDDL(withCache)
	if err != nil {
		return err
	}

	for tableName, code := range modelList {
		name := fmt.Sprintf("%s.go", strings.ToLower(stringx.From(tableName).ToSnake()))
		filename := filepath.Join(dir, name)
		//if fs.FileExist(filename) {
		//	g.Warning("%s 已存在，跳过。", name)
		//	continue
		//}
		err = ioutil.WriteFile(filename, []byte(code), os.ModePerm)
		if err != nil {
			fmt.Println("生成出错")
			fmt.Println(filename)
			fmt.Println(code)
			fmt.Println(err)
			return err
		}
	}

	filename := filepath.Join(dir, "vars.go")
	if !fs.FileExist(filename) {
		err = ioutil.WriteFile(filename, []byte(tpl.Error), os.ModePerm)
		if err != nil {
			return err
		}
	}

	g.Success("完成。")
	return nil
}

func (g *ModelGenerator) genFromDDL(withCache bool) (map[string]string, error) {
	m := make(map[string]string)
	for _, ddl := range g.ddlList {
		table, err := parser.Parse(ddl)
		if err != nil {
			return nil, err
		}
		modelCode, err := g.genModelCode(*table, withCache)
		if err != nil {
			return nil, err
		}
		m[table.Name.Source()] = modelCode
	}
	return m, nil
}

func (g *ModelGenerator) genModelCode(table parser.Table, withCache bool) (string, error) {
	// 生成缓存键代码
	cacheKeys, err := genCacheKeys(g.database, table)
	if err != nil {
		return "", err
	}

	var tableDTO Table
	tableDTO.Table = table
	tableDTO.CacheKeys = cacheKeys

	// 生成导包代码
	importsCode, err := genImports(withCache, tableDTO.ContainsTime())
	if err != nil {
		return "", nil
	}

	// 生成变量声明代码段
	varsCode, err := genVars(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成类型声明代码段
	typesCode, err := genTypes(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成新生成模型的代码段
	newCode, err := genNew(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成数据插入代码段
	insertCode, err := genInsert(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成事务型数据插入代码段
	txInsertCode, err := genTxInsert(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成主键查找代码段
	findOneCode, err := genFindOne(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成一组主键查找代码段
	findManyCode, err := genFindMany(tableDTO)
	if err != nil {
		return "", nil
	}

	// 生成唯一键字段查找代码段
	findOneByFieldCode, err := genFindOneByField(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	findManyByFieldsCode, err := genFindManyByFields(tableDTO)
	if err != nil {
		return "", err
	}

	// 合成查找代码段
	findCode := make([]string, 0)
	findCode = append(findCode, findOneCode, findManyCode, findOneByFieldCode, findManyByFieldsCode)

	// 生成更新代码段
	updateCode, err := genUpdate(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成局部更新代码段
	updatePartialCode, err := genUpdatePartial(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成事务型更新代码段
	txUpdateCode, err := genTxUpdate(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成事务型局部更新代码段
	txUpdatePartialCode, err := genTxUpdatePartial(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 生成删除代码段
	deleteCode, err := genDelete(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 合成删除代码段
	txDeleteCode, err := genTxDelete(tableDTO, withCache)
	if err != nil {
		return "", nil
	}

	// 合成并输出模板字符串
	output, err := util.With("model").
		Parse(tpl.Model).
		GoFmt(true).
		Execute(map[string]interface{}{
			"imports":         importsCode,
			"vars":            varsCode,
			"types":           typesCode,
			"new":             newCode,
			"insert":          insertCode,
			"txInsert":        txInsertCode,
			"find":            strings.Join(findCode, "\n"),
			"update":          updateCode,
			"updatePartial":   updatePartialCode,
			"txUpdate":        txUpdateCode,
			"txUpdatePartial": txUpdatePartialCode,
			"delete":          deleteCode,
			"txDelete":        txDeleteCode,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}
