package template

// Error 定义一个变量模板。
const Error = `package {{.pkg}}

import "github.com/gotid/god/lib/store/sqlx"

var ErrNotFound = sqlx.ErrNotFound
`
