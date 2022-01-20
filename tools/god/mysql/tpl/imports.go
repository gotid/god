package tpl

var (
	Imports = `import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	{{if .time}}"time"{{end}}

	"github.com/gotid/god/lib/container/garray"
	"github.com/gotid/god/lib/gconv"
	"github.com/gotid/god/lib/fx"
	"github.com/gotid/god/lib/gutil"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/mr"
	"github.com/gotid/god/lib/g"
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/tools/god/mysql/builder"
)
`

	ImportsNoCache = `import (
	"database/sql"
	"sort"
	"strings"
	{{if .time}}"time"{{end}}

	"github.com/gotid/god/lib/container/garray"
	"github.com/gotid/god/lib/gconv"
	"github.com/gotid/god/lib/gutil"
	"github.com/gotid/god/lib/mathx"
	"github.com/gotid/god/lib/mr"
	"github.com/gotid/god/lib/g"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/lib/stringx"
	"github.com/gotid/god/tools/god/mysql/builder"
)`
)
