package internal

import (
	"bytes"
	"fmt"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/rest/httpx"
	"net/http"
	"sync"
)

type contextKey string

func (c contextKey) String() string {
	return "rest/internal context key " + string(c)
}

// LogContext 是一个上下文键。
var LogContext = contextKey("request_logs")

// LogCollector 用于收集日志。
type LogCollector struct {
	Messages []string
	lock     sync.Mutex
}

// Append 追加内容到日志上下文。
func (c *LogCollector) Append(msg string) {
	c.lock.Lock()
	c.Messages = append(c.Messages, msg)
	c.lock.Unlock()
}

// Flush 刷新日志。将已收集的日志并汇集成一条字符串。
func (c *LogCollector) Flush() string {
	var buf bytes.Buffer

	start := true
	for _, msg := range c.takeAll() {
		if start {
			start = false
		} else {
			buf.WriteByte('\n')
		}
		buf.WriteString(msg)
	}

	return buf.String()
}

func (c *LogCollector) takeAll() []string {
	c.lock.Lock()
	messages := c.Messages
	c.Messages = nil
	c.lock.Unlock()

	return messages
}

// Error 将 http 请求和给定的变量一起写入错误日志。
func Error(r *http.Request, v ...interface{}) {
	logx.WithContext(r.Context()).Error(format(r, v...))
}

// Errorf 将 http 请求和变量以给定的格式给一起写入错误日志。
func Errorf(r *http.Request, format string, v ...interface{}) {
	logx.WithContext(r.Context()).Error(formatf(r, format, v...))
}

// Info 将 http 请求和给定的变量一起写如访问日志。
func Info(r *http.Request, v ...interface{}) {
	appendLog(r, format(r, v...))
}

// Infof 将 http 请求和变量以给定的格式给一起写如访问日志。
func Infof(r *http.Request, format string, v ...interface{}) {
	appendLog(r, formatf(r, format, v...))
}

func appendLog(r *http.Request, msg string) {
	logs := r.Context().Value(LogContext)
	if logs != nil {
		logs.(*LogCollector).Append(msg)
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
