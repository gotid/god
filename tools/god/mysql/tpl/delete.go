package tpl

var Delete = `
func (m *{{.upperStartCamelObject}}Model) Delete({{.lowerStartCamelPrimaryKey}} ...{{.dataType}}) error {
	if len({{.lowerStartCamelPrimaryKey}}) == 0 {
		return nil
	}

	{{if .withCache}}{{if .containsIndexCache}}datas:=m.FindMany({{.lowerStartCamelPrimaryKey}}){{end}}
	{{.keys}}

    _, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := fmt.Sprintf(` + "`" + `delete from ` + "` +" + ` m.table + ` + " `" + ` where {{.originalPrimaryKey}} in (%s)` + "`" + `, sqlx.In(len(id)))
		return conn.Exec(query, gconv.Interfaces({{.lowerStartCamelPrimaryKey}})...)
	}, keys...){{else}}query := ` + "`" + `delete from ` + "` +" + ` m.table + ` + " `" + ` where {{.originalPrimaryKey}} = ?` + "`" + `
		_,err:=m.conn.Exec(query, {{.lowerStartCamelPrimaryKey}}){{end}}
	return err
}
`

var TxDelete = `
func (m *{{.upperStartCamelObject}}Model) TxDelete(tx sqlx.TxSession, {{.lowerStartCamelPrimaryKey}} ...{{.dataType}}) error {
	if len({{.lowerStartCamelPrimaryKey}}) == 0 {
		return nil
	}

	{{if .withCache}}{{if .containsIndexCache}}datas:=m.FindMany({{.lowerStartCamelPrimaryKey}}){{end}}
	{{.keys}}

    _, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := fmt.Sprintf(` + "`" + `delete from ` + "` +" + ` m.table + ` + " `" + ` where {{.originalPrimaryKey}} in (%s)` + "`" + `, sqlx.In(len(id)))
		return tx.Exec(query, gconv.Interfaces({{.lowerStartCamelPrimaryKey}})...)
	}, keys...){{else}}query := ` + "`" + `delete from ` + "` +" + ` m.table + ` + " `" + ` where {{.originalPrimaryKey}} = ?` + "`" + `
		_,err := tx.Exec(query, {{.lowerStartCamelPrimaryKey}}){{end}}
	return err
}
`
