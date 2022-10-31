package template

const (
	// Insert 定义一个模型中的插入代码模板。
	Insert = `
func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, data *{{.upperStartCamelObject}}) (sql.Result,error) {
	{{if .withCache}}{{.keys}}
    ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.Conn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values ({{.expression}})", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, {{.expressionValues}})
	}, {{.keyValues}}){{else}}query := fmt.Sprintf("insert into %s (%s) values ({{.expression}})", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
    ret, err:=m.conn.ExecCtx(ctx, query, {{.expressionValues}}){{end}}
	return ret, err
}
`

	// InsertMethod 定义一个用于模型中插入代码的接口方法模板。
	InsertMethod = `Insert(ctx context.Context, data *{{.upperStartCamelObject}}) (sql.Result, error)`
)
