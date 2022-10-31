package template

import (
	"fmt"
	"github.com/gotid/god/tools/god/util"
)

const ModelCustom = `package {{.pkg}}
{{if .withCache}}
import (
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/sqlx"
)
{{else}}
import "github.com/gotid/god/lib/store/sqlx"
{{end}}
`

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
