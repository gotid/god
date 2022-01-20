package g

import (
	"github.com/gotid/god/internal/empty"
	"github.com/gotid/god/lib/container/gvar"
	"github.com/gotid/god/lib/gutil"
)

// NewVar returns a gvar.Var.
func NewVar(i interface{}, safe ...bool) *Var {
	return gvar.New(i, safe...)
}

// Dump dumps a variable to stdout with more manually readable.
func Dump(i ...interface{}) {
	gutil.Dump(i...)
}

// Export exports a variable to string with more manually readable.
func Export(i ...interface{}) string {
	return gutil.Export(i...)
}

// Throw throws a exception, which can be caught by TryCatch function.
// It always be used in TryCatch function.
func Throw(exception interface{}) {
	gutil.Throw(exception)
}

// Try implements try... logistics using internal panic...recover.
// It returns error if any exception occurs, or else it returns nil.
func Try(try func()) (err error) {
	return gutil.Try(try)
}

// TryCatch implements try...catch... logistics using internal panic...recover.
// It automatically calls function <catch> if any exception occurs ans passes the exception as an error.
func TryCatch(try func(), catch ...func(exception error)) {
	gutil.TryCatch(try, catch...)
}

// IsNil checks whether given <value> is nil.
// Note that it might use reflect feature which affects performance a little bit.
func IsNil(value interface{}) bool {
	return empty.IsNil(value)
}

// IsEmpty checks whether given <value> empty.
// It returns true if <value> is in: 0, nil, false, "", len(slice/map/chan) == 0.
// Or else it returns true.
func IsEmpty(value interface{}) bool {
	return empty.IsEmpty(value)
}
