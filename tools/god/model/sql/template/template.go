package template

import (
	_ "embed"
	"fmt"
	"github.com/gotid/god/tools/god/util"
)

// Vars 定义模型中变量区块的模板。
//
//go:embed tpl/vars.tpl
var Vars string

// Types 定义模型中用到的类型声明模板。
//
//go:embed tpl/types.tpl
var Types string

// Tag 定义一个 tag 模板文本。
//
//go:embed tpl/tag.tpl
var Tag string

// TableName 定义生成表名方法的模板。
//
//go:embed tpl/table-name.tpl
var TableName string

// New 定义用于创建模型实例的模板。
//
//go:embed tpl/model-new.tpl
var New string

// ModelCustom 定义创建模型实例的代码模板。
//
//go:embed tpl/model.tpl
var ModelCustom string

// ModelGen 定义一个模型的模板。
var ModelGen = fmt.Sprintf(`%s

package {{.pkg}}
{{.imports}}
{{.vars}}
{{.types}}
{{.new}}
{{.insert}}
{{.delete}}
{{.update}}
{{.find}}
{{.extraMethod}}
{{.tableName}}
`, util.DontEditHead)

// Insert 定义一个模型中的插入代码模板。
//
//go:embed tpl/insert.tpl
var Insert string

// InsertMethod 定义一个用于模型中插入代码的接口方法模板。
//
//go:embed tpl/interface-insert.tpl
var InsertMethod string

// Update 定义一个生成更新代码的模板。
//
//go:embed tpl/update.tpl
var Update string

// UpdateMethod 定义生成更新代码的接口方法模板。
//
//go:embed tpl/interface-update.tpl
var UpdateMethod string

// Imports 定义有缓存场景的模型导入模板。
//
//go:embed tpl/import-with-cache.tpl
var Imports string

// ImportsNoCache 定义无缓存场景的模型导入模板。
//
//go:embed tpl/import-no-cache.tpl
var ImportsNoCache string

// FindOne 根据 id 找单行。
//
//go:embed tpl/find-one.tpl
var FindOne string

// FindOneByField 根据字段找单行。
//
//go:embed tpl/find-one-by-field.tpl
var FindOneByField string

// FindOneByFieldExtraMethod 根据字段和扩展信息找单行。
//
//go:embed tpl/find-one-by-field-extra-method.tpl
var FindOneByFieldExtraMethod string

// FindOneMethod 定义找单行的方法。
//
//go:embed tpl/interface-find-one.tpl
var FindOneMethod string

// FindOneByFieldMethod 定义根据字段找单行的方法。
//
//go:embed tpl/interface-find-one-by-field.tpl
var FindOneByFieldMethod string

// Field 定义一个字段类型声明的模板。
//
//go:embed tpl/field.tpl
var Field string

// Error 定义一个错误模板。
//
//go:embed tpl/err.tpl
var Error string

// Delete 定义删除代码的模板
//
//go:embed tpl/delete.tpl
var Delete string

// DeleteMethod 定义删除代码的模板方法
//
//go:embed tpl/interface-delete.tpl
var DeleteMethod string
