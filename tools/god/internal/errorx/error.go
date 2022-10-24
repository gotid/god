package errorx

import (
	"fmt"
	"github.com/gotid/god/tools/god/pkg/env"
	"strings"
)

var errFormat = `god error: %+v
god env:
%s
%s`

// GodErr 表示一个 God 代码生成器错误。
type GodErr struct {
	msg []string
	err error
}

func (e *GodErr) Error() string {
	detail := wrapMsg(e.msg...)
	return fmt.Sprintf(errFormat, e.err, env.Print(), detail)
}

func wrapMsg(msg ...string) string {
	if len(msg) == 0 {
		return ""
	}

	return fmt.Sprintf(`消息：%s`, strings.Join(msg, "\n"))
}

// Wrap 用 god 版本和给定消息包装一个错误。
func Wrap(err error, msg ...string) error {
	e, ok := err.(*GodErr)
	if ok {
		return e
	}

	return &GodErr{
		msg: msg,
		err: err,
	}
}
