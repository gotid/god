package template

import _ "embed"

// Vars 定义模型中变量区块的模板。
//
//go:embed vars.tpl
var Vars string
