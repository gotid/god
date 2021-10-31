package logx

import (
	"context"
	"fmt"
	"io"
	"time"

	"git.zc0901.com/go/god/lib/timex"
	"go.opentelemetry.io/otel/trace"
)

type traceLogger struct {
	logEntry
	Trace string `json:"trace,omitempty"`
	Span  string `json:"span,omitempty"`
	ctx   context.Context
}

func (l *traceLogger) Info(v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprint(v...))
	}
}

func (l *traceLogger) Infof(format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprintf(format, v...))
	}
}

func (l *traceLogger) Infov(v interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, v)
	}
}

func (l *traceLogger) Error(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *traceLogger) Errorf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *traceLogger) Errorv(v interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, v)
	}
}

func (l *traceLogger) Slow(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprint(v...))
	}
}

func (l *traceLogger) Slowf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprintf(format, v...))
	}
}

func (l *traceLogger) Slowv(v interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, v)
	}
}

func (l *traceLogger) WithDuration(duration time.Duration) Logger {
	l.Duration = timex.MillisecondDuration(duration)
	return l
}

func (l *traceLogger) write(writer io.WriteCloser, level string, content interface{}) {
	l.Timestamp = getTimestamp()
	l.Level = level
	l.Content = content
	l.Trace = traceIdFromContext(l.ctx)
	l.Span = spanIdFromContext(l.ctx)
	outputJson(writer, l)
}

func WithContext(ctx context.Context) Logger {
	return &traceLogger{
		ctx: ctx,
	}
}

func traceIdFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func spanIdFromContext(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}
