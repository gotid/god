package template

// TableName 定义生成表名方法的模板。
const TableName = `
func (m *default{{.upperStartCamelObject}}Model) tableName() string {
	return m.table
}
`
