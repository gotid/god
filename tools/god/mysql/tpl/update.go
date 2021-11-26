package tpl

var Update = `
func (m *{{.upperTable}}Model) Update(data {{.upperTable}}) error {
	{{if .withCache}}{{.primaryCacheKey}}
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := ` + "`" + `update ` + "` +" + ` m.table +` + "` " + `set ` + "` + " + `{{.lowerTable}}FieldsWithPlaceHolder` + " + `" + ` where {{.originalPrimaryKey}} = ?` + "`" + `
		return conn.Exec(query, {{.values}})
	}, {{.primaryKeyName}}){{else}}query := ` + "`" + `update ` + "` +" + `m.table +` + "` " + `set ` + "` +" + `{{.lowerTable}}FieldsWithPlaceHolder` + " + `" + ` where {{.originalPrimaryKey}} = ?` + "`" + `
	_,err := m.conn.Exec(query, {{.values}}){{end}}
	return err
}
`

var UpdatePartial = `
func (m *{{.upperTable}}Model) UpdatePartial(ms ...g.Params) (err error) {
	okNum := 0
	fx.From(func(source chan<- interface{}) {
		for _, data := range ms {
			source <- data
		}
	}).Parallel(func(item interface{}) {
		err = m.updatePartial(item.(g.Params))
		if err != nil {
			return
		}
		okNum++
	})

	if err == nil && okNum != len(ms) {
		err = fmt.Errorf("部分局部更新失败！待更新(%d) != 实际更新(%d)", len(ms), okNum)
	}

	return err
}

func (m *{{.upperTable}}Model) updatePartial(data g.Params) error {
	updateArgs, err := sqlx.ExtractUpdateArgs({{.lowerTable}}FieldList, data)
	if err != nil {
		return err
	}

	{{if .withCache}}{{.primaryCacheKey}}
	_, err = m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := ` + "`" + `update ` + "` +" + ` m.table +` + "` " + `set ` + "` + " + `updateArgs.Fields` + " + `" + ` where {{.originalPrimaryKey}} = ` + "` + " + `updateArgs.Id` + `
		return conn.Exec(query, updateArgs.Args...)
	}, {{.primaryKeyName}}){{else}}query := ` + "`" + `update ` + "` +" + `m.table +` + "` " + `set ` + "` +" + `updateArgs.Fields` + " + `" + ` where {{.originalPrimaryKey}} = ` + "` + " + `updateArgs.Id` + `
	_,err = m.conn.Exec(query, updateArgs.Args...){{end}}
	return err
}
`

var TxUpdate = `
func (m *{{.upperTable}}Model) TxUpdate(tx sqlx.TxSession, data {{.upperTable}}) error {
	{{if .withCache}}{{.primaryCacheKey}}
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := ` + "`" + `update ` + "` +" + ` m.table +` + "` " + `set ` + "` + " + `{{.lowerTable}}FieldsWithPlaceHolder` + " + `" + ` where {{.originalPrimaryKey}} = ?` + "`" + `
		return tx.Exec(query, {{.values}})
	}, {{.primaryKeyName}}){{else}}query := ` + "`" + `update ` + "` +" + `m.table +` + "` " + `set ` + "` +" + `{{.lowerTable}}FieldsWithPlaceHolder` + " + `" + ` where {{.originalPrimaryKey}} = ?` + "`" + `
	_,err := tx.Exec(query, {{.values}}){{end}}
	return err
}
`

var TxUpdatePartial = `
func (m *{{.upperTable}}Model) TxUpdatePartial(tx sqlx.TxSession, ms ...g.Params) (err error) {
	okNum := 0
	fx.From(func(source chan<- interface{}) {
		for _, data := range ms {
			source <- data
		}
	}).Parallel(func(item interface{}) {
		err = m.txUpdatePartial(tx, item.(g.Params))
		if err != nil {
			return
		}
		okNum++
	})

	if err == nil && okNum != len(ms) {
		err = fmt.Errorf("部分事务型局部更新失败！待更新(%d) != 实际更新(%d)", len(ms), okNum)
	}
	return err
}

func (m *{{.upperTable}}Model) txUpdatePartial(tx sqlx.TxSession, data g.Params) error {
	updateArgs, err := sqlx.ExtractUpdateArgs({{.lowerTable}}FieldList, data)
	if err != nil {
		return err
	}

	{{if .withCache}}{{.primaryCacheKey}}
	_, err = m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := ` + "`" + `update ` + "` +" + ` m.table +` + "` " + `set ` + "` + " + `updateArgs.Fields` + " + `" + ` where {{.originalPrimaryKey}} = ` + "` + " + `updateArgs.Id` + `
		return tx.Exec(query, updateArgs.Args...)
	}, {{.primaryKeyName}}){{else}}query := ` + "`" + `update ` + "` +" + `m.table +` + "` " + `set ` + "` +" + `updateArgs.Fields` + " + `" + ` where {{.originalPrimaryKey}} = ` + "` + " + `updateArgs.Id` + `
	_,err = tx.Exec(query, updateArgs.Args...){{end}}
	return err
}
`
