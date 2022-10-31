package template

// Field 定义一个字段类型声明的模板。
const Field = `{{.name}} {{.type}} {{.tag}} {{if .hasComment}}// {{.comment}}{{end}}`
