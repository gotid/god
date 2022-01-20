package internal

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/lib/logx"
)

// LogContext 是一个上下文键。
var LogContext = contextKey("request_logs")

// LogCollector 是一个日志采集器。
type LogCollector struct {
	Messages []string
	lock     sync.Mutex
}

// Append 追加消息到日志上下文。
func (c *LogCollector) Append(msg string) {
	c.lock.Lock()
	c.Messages = append(c.Messages, msg)
	c.lock.Unlock()
}

// Flush 刷新已收集的日志。
func (c *LogCollector) Flush() string {
	var b bytes.Buffer

	start := true
	for _, msg := range c.takeAll() {
		if start {
			start = false
		} else {
			b.WriteByte('\n')
		}
		b.WriteString(msg)
	}

	return b.String()
}

func (c *LogCollector) takeAll() []string {
	c.lock.Lock()
	messages := c.Messages
	c.Messages = nil
	c.lock.Unlock()

	return messages
}

// Error 将请求和错误写入错误日志。
func Error(r *http.Request, v ...interface{}) {
	logx.ErrorCaller(1, format(r, v...))
}

// Errorf 将请求和错误按指定格式写入错误日志。
func Errorf(r *http.Request, format string, v ...interface{}) {
	logx.ErrorCaller(1, formatf(r, format, v...))
}

// Info 将请求和信息写入访问日志。
func Info(r *http.Request, v ...interface{}) {
	appendLog(r, format(r, v...))
}

// Infof 将请求和信息按指定格式写入访问日志。
func Infof(r *http.Request, format string, v ...interface{}) {
	appendLog(r, formatf(r, format, v...))
}

func appendLog(r *http.Request, message string) {
	logs := r.Context().Value(LogContext)
	if logs != nil {
		logs.(*LogCollector).Append(message)
	}
}

func format(r *http.Request, v ...interface{}) string {
	return formatWithReq(r, fmt.Sprint(v...))
}

func formatf(r *http.Request, format string, v ...interface{}) string {
	return formatWithReq(r, fmt.Sprintf(format, v...))
}

func formatWithReq(r *http.Request, v string) string {
	return fmt.Sprintf("(%s - %s) %s", r.RequestURI, httpx.GetRemoteAddr(r), v)
}

type contextKey string

func (c contextKey) String() string {
	return "api/internal pathvar key " + string(c)
}
