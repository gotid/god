package tpl

var (
	Imports = `import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	{{if .time}}"time"{{end}}

	"git.zc0901.com/go/god/lib/container/garray"
	"git.zc0901.com/go/god/lib/gconv"
	"git.zc0901.com/go/god/lib/fx"
	"git.zc0901.com/go/god/lib/gutil"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/mathx"
	"git.zc0901.com/go/god/lib/mr"
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/store/cache"
	"git.zc0901.com/go/god/lib/store/sqlx"
	"git.zc0901.com/go/god/lib/stringx"
	"git.zc0901.com/go/god/tools/god/mysql/builder"
)
`

	ImportsNoCache = `import (
	"database/sql"
	"sort"
	"strings"
	{{if .time}}"time"{{end}}

	"git.zc0901.com/go/god/lib/container/garray"
	"git.zc0901.com/go/god/lib/gconv"
	"git.zc0901.com/go/god/lib/gutil"
	"git.zc0901.com/go/god/lib/mathx"
	"git.zc0901.com/go/god/lib/mr"
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/store/sqlx"
	"git.zc0901.com/go/god/lib/stringx"
	"git.zc0901.com/go/god/tools/god/mysql/builder"
)`
)
