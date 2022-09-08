package logx

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	fatihcolor "github.com/fatih/color"
	"github.com/gotid/god/core/color"
)

type (
	// Writer 编写器接口。
	Writer interface {
		Close() error
		Info(v interface{}, fields ...LogField)
		Alert(v interface{})
		Error(v interface{}, fields ...LogField)
		Severe(v interface{})
		Slow(v interface{}, fields ...LogField)
		Stack(v interface{})
		Stat(v interface{}, fields ...LogField)
	}

	// 原子性编写器，会调用具体编写器 concreteWriter。
	atomicWriter struct {
		writer Writer
		lock   sync.RWMutex
	}

	// 具体的日志编写器。
	concreteWriter struct {
		infoLog   io.WriteCloser
		errorLog  io.WriteCloser
		severeLog io.WriteCloser
		slowLog   io.WriteCloser
		statLog   io.WriteCloser
		stackLog  io.Writer
	}

	// 无操作的日志编写器。
	nopWriter struct{}
)

// Load 加载当前日志编写器。
func (w *atomicWriter) Load() Writer {
	w.lock.RLock()
	defer w.lock.RUnlock()
	return w.writer
}

// Store 将 v 保存为当前日志编写器。
func (w *atomicWriter) Store(v Writer) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.writer = v
}

// StoreIfNil 当前如无日志编写器则保存。
func (w *atomicWriter) StoreIfNil(v Writer) Writer {
	w.lock.Lock()
	defer w.lock.Unlock()

	if w.writer == nil {
		w.writer = v
	}

	return w.writer
}

// Swap 将 v 保存为当前编写器，并返回原编写器。
func (w *atomicWriter) Swap(v Writer) Writer {
	w.lock.Lock()
	defer w.lock.Unlock()

	old := w.writer
	w.writer = v
	return old
}

// Close 关闭具体的日志编写器。
func (w *concreteWriter) Close() error {
	if err := w.infoLog.Close(); err != nil {
		return err
	}
	if err := w.errorLog.Close(); err != nil {
		return err
	}
	if err := w.severeLog.Close(); err != nil {
		return err
	}
	if err := w.slowLog.Close(); err != nil {
		return err
	}
	return w.statLog.Close()
}

// Info 调用具体编写器进行通知。
func (w *concreteWriter) Info(v interface{}, fields ...LogField) {
	output(w.infoLog, levelInfo, v, fields...)
}

// Alert 调用具体编写器进行警告。
func (w *concreteWriter) Alert(v interface{}) {
	output(w.errorLog, levelAlert, v)
}

// Error 调用具体编写器进行错误告警。
func (w *concreteWriter) Error(v interface{}, fields ...LogField) {
	output(w.errorLog, levelError, v, fields...)
}

// Severe 调用具体编写器进行严重告警。
func (w *concreteWriter) Severe(v interface{}) {
	output(w.severeLog, levelFatal, v)
}

// Slow 调用具体编写器进行慢执行记录。
func (w *concreteWriter) Slow(v interface{}, fields ...LogField) {
	output(w.slowLog, levelSlow, v, fields...)
}

// Stack 调用具体编写器进行错误堆栈记录。
func (w *concreteWriter) Stack(v interface{}) {
	output(w.stackLog, levelError, v)
}

// Stat 调用具体编写器进行统计记录。
func (w *concreteWriter) Stat(v interface{}, fields ...LogField) {
	output(w.statLog, levelStat, v, fields...)
}

func (n nopWriter) Close() error {
	return nil
}

func (n nopWriter) Alert(_ interface{})                {}
func (n nopWriter) Error(_ interface{}, _ ...LogField) {}
func (n nopWriter) Info(_ interface{}, _ ...LogField)  {}
func (n nopWriter) Severe(_ interface{})               {}
func (n nopWriter) Slow(_ interface{}, _ ...LogField)  {}
func (n nopWriter) Stack(_ interface{})                {}
func (n nopWriter) Stat(_ interface{}, _ ...LogField)  {}

// NewWriter 创建并返回一个给定的编写器对应的具体编写器。
func NewWriter(writer io.Writer) Writer {
	lw := newLogWriter(log.New(writer, "", flags))

	return &concreteWriter{
		infoLog:   lw,
		errorLog:  lw,
		severeLog: lw,
		slowLog:   lw,
		statLog:   lw,
		stackLog:  lw,
	}
}

// 返回一个控制台编写器作为具体编写器。
func newConsoleWriter() Writer {
	outLog := newLogWriter(log.New(fatihcolor.Output, "", flags))
	errLog := newLogWriter(log.New(fatihcolor.Error, "", flags))
	return &concreteWriter{
		infoLog:   outLog,
		errorLog:  errLog,
		severeLog: errLog,
		slowLog:   errLog,
		stackLog:  newLessWriter(errLog, options.logStackCooldownMillis),
		statLog:   outLog,
	}
}

// 返回一个文件编写器作为具体编写器。
func newFileWriter(c LogConf) (Writer, error) {
	var (
		err       error
		opts      []LogOption
		infoLog   io.WriteCloser
		errorLog  io.WriteCloser
		severeLog io.WriteCloser
		slowLog   io.WriteCloser
		statLog   io.WriteCloser
		stackLog  io.Writer
	)

	if len(c.Path) == 0 {
		return nil, ErrLogPathNotSet
	}

	opts = append(opts, WithCooldownMillis(c.StackCooldownMillis))
	if c.Compress {
		opts = append(opts, WithGzip())
	}
	if c.KeepDays > 0 {
		opts = append(opts, WithKeepDays(c.KeepDays))
	}
	if c.MaxBackups > 0 {
		opts = append(opts, WithMaxBackups(c.MaxBackups))
	}
	if c.MaxSize > 0 {
		opts = append(opts, WithMaxSize(c.MaxSize))
	}

	opts = append(opts, WithRotation(c.Rotation))

	accessFile := path.Join(c.Path, accessFilename)
	errorFile := path.Join(c.Path, errorFilename)
	severeFile := path.Join(c.Path, severeFilename)
	slowFile := path.Join(c.Path, slowFilename)
	statFile := path.Join(c.Path, statFilename)

	handleOptions(opts)
	setupLogLevel(c)

	if infoLog, err = createOutput(accessFile); err != nil {
		return nil, err
	}
	if errorLog, err = createOutput(errorFile); err != nil {
		return nil, err
	}
	if severeLog, err = createOutput(severeFile); err != nil {
		return nil, err
	}
	if slowLog, err = createOutput(slowFile); err != nil {
		return nil, err
	}
	if statLog, err = createOutput(statFile); err != nil {
		return nil, err
	}

	stackLog = newLessWriter(errorLog, options.logStackCooldownMillis)

	return &concreteWriter{
		infoLog:   infoLog,
		errorLog:  errorLog,
		severeLog: severeLog,
		slowLog:   slowLog,
		statLog:   statLog,
		stackLog:  stackLog,
	}, nil
}

func output(writer io.Writer, level string, val interface{}, fields ...LogField) {
	fields = append(fields, Field(callerKey, getCaller(callerDepth)))

	switch atomic.LoadUint32(&encoding) {
	case plainEncodingType:
		writePlainAny(writer, level, val, buildFields(fields...)...)
	default:
		entry := make(logEntryWithFields)
		for _, field := range fields {
			entry[field.Key] = field.Value
		}
		entry[timestampKey] = getTimestamp()
		entry[levelKey] = level
		entry[contentKey] = val
		writeJson(writer, entry)
	}
}

func writeJson(writer io.Writer, info interface{}) {
	if content, err := json.Marshal(info); err != nil {
		log.Println(err.Error())
	} else if writer == nil {
		log.Println(string(content))
	} else {
		writer.Write(append(content, '\n'))
	}
}

func writePlainAny(writer io.Writer, level string, val interface{}, fields ...string) {
	level = wrapLevelWithColor(level)

	switch v := val.(type) {
	case string:
		writePlainText(writer, level, v, fields...)
	case error:
		writePlainText(writer, level, v.Error(), fields...)
	case fmt.Stringer:
		writePlainText(writer, level, v.String(), fields...)
	default:
		writePlainValue(writer, level, v, fields...)
	}
}

func writePlainText(writer io.Writer, level, msg string, fields ...string) {
	var buf strings.Builder
	buf.WriteString(getTimestamp())
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(level)
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(msg)
	for _, item := range fields {
		buf.WriteByte(plainEncodingSep)
		buf.WriteString(item)
	}
	buf.WriteByte('\n')
	if writer == nil {
		log.Println(buf.String())
		return
	}

	if _, err := fmt.Fprint(writer, buf.String()); err != nil {
		log.Println(err.Error())
	}
}

func writePlainValue(writer io.Writer, level string, v interface{}, fields ...string) {
	var buf strings.Builder
	buf.WriteString(getTimestamp())
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(level)
	buf.WriteByte(plainEncodingSep)
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Println(err.Error())
		return
	}

	for _, item := range fields {
		buf.WriteByte(plainEncodingSep)
		buf.WriteString(item)
	}
	buf.WriteByte('\n')
	if writer == nil {
		log.Println(buf.String())
		return
	}

	if _, err := fmt.Fprint(writer, buf.String()); err != nil {
		log.Println(err.Error())
	}
}

// 返回上色后的日志级别
func wrapLevelWithColor(level string) string {
	var colour color.Color
	switch level {
	case levelAlert:
		colour = color.FgRed
	case levelError:
		colour = color.FgRed
	case levelSevere:
		colour = color.FgRed
	case levelFatal:
		colour = color.FgRed
	case levelInfo:
		colour = color.FgBlue
	case levelSlow:
		colour = color.FgYellow
	case levelStat:
		colour = color.FgGreen
	}

	if colour == color.NoColor {
		return level
	}

	return color.WithColorPadding(level, colour)
}

func buildFields(fields ...LogField) []string {
	var items []string

	for _, field := range fields {
		items = append(items, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}

	return items
}
