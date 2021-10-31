package logx

import (
	"fmt"
	"io"
	"time"

	"git.zc0901.com/go/god/lib/timex"
)

const durationCallerDepth = 3

type durationLogger logEntry

// WithDuration 返回一个指定时长的日志。
func WithDuration(d time.Duration) Logger {
	return &durationLogger{
		Duration: timex.MillisecondDuration(d),
	}
}

func (l *durationLogger) Info(v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Infof(format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Infov(v interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, v)
	}
}

func (l *durationLogger) Error(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *durationLogger) Errorf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *durationLogger) Errorv(v interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, v)
	}
}

func (l *durationLogger) Slow(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprint(v...))
	}
}

func (l *durationLogger) Slowf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprintf(format, v...))
	}
}

func (l *durationLogger) Slowv(v interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, v)
	}
}

func (l *durationLogger) WithDuration(d time.Duration) Logger {
	l.Duration = timex.MillisecondDuration(d)
	return l
}

func (l *durationLogger) write(writer io.Writer, level string, val interface{}) {
	l.Timestamp = getTimestamp()
	l.Level = level
	l.Content = val
	outputJson(writer, l)
}
