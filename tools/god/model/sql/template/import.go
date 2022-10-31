package template

const (
	// Imports 定义缓存场景的模型导入模板
	Imports = `import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	{{if .containsPQ}}"github.com/lib/pq"{{end}}
	"github.com/gotid/god/lib/store/builder"
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/sqlc"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/lib/stringx"
)
`
	// ImportsNoCache 定义常规无缓存场景的模型导入模板
	ImportsNoCache = `import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	{{if .time}}"time"{{end}}

	{{if .containsPQ}}"github.com/lib/pq"{{end}}
	"github.com/gotid/god/lib/store/builder"
	"github.com/gotid/god/lib/store/sqlc"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/lib/stringx"
)
`
)
