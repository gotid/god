package logx

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gotid/god/lib/sysx"
)

const callerDepth = 4

var (
	timeFormat  = "2006-01-02T15:04:05.000Z07:00"
	logLevel    uint32
	encoding    uint32 = jsonEncodingType
	writer             = new(atomicWriter)
	disableLog  uint32
	disableStat uint32
	options     logOptions
	setupOnce   sync.Once
)

type (
	// LogField 是一个将被加入到日志条目的键值对。
	LogField struct {
		Key   string
		Value any
	}

	// LogOption 用于自定义日志记录的方法。
	LogOption func(options *logOptions)

	logEntry map[string]any

	logOptions struct {
		gzipEnabled            bool
		logStackCooldownMillis int
		keepDays               int
		maxBackups             int
		maxSize                int
		rotationRule           string
	}
)

// MustSetup 使用给定配置 c 进行设置。有错则退出。
func MustSetup(c Config) {
	Must(Setup(c))
}

// Must 检查是否有错，有错则记录错误并退出。
func Must(err error) {
	if err == nil {
		return
	}

	msg := err.Error()
	log.Print(msg)
	getWriter().Severe(msg)
	os.Exit(1)
}

// Setup 设置 logx。
// 允许在不同服务框架中多次调用，但只设置一次。
func Setup(c Config) (err error) {
	setupOnce.Do(func() {
		setupLogLevel(c)

		if len(c.TimeFormat) > 0 {
			timeFormat = c.TimeFormat
		}

		switch c.Encoding {
		case plainEncoding:
			atomic.StoreUint32(&encoding, plainEncodingType)
		default:
			atomic.StoreUint32(&encoding, jsonEncodingType)
		}

		switch c.Mode {
		case fileMode:
			err = setupWithFiles(c)
		case volumeMode:
			err = setupWithVolume(c)
		default:
			setupWithConsole()
		}
	})

	return
}

// Close 关闭日志记录。
func Close() error {
	if w := writer.Swap(nil); w != nil {
		return w.(io.Closer).Close()
	}

	return nil
}

// Disable 禁用日志记录。
func Disable() {
	atomic.StoreUint32(&disableLog, 1)
	writer.Store(nopWriter{})
}

// DisableStat 禁用统计记录。
func DisableStat() {
	atomic.StoreUint32(&disableStat, 1)
}

// SetWriter 设置日志编写器，用于自定义日志记录行为。
func SetWriter(w Writer) {
	if atomic.LoadUint32(&disableLog) == 0 {
		writer.Store(w)
	}
}

// Reset 清空当前编写器，返回原编写器。
func Reset() Writer {
	return writer.Swap(nil)
}

// SetLevel 设置记录级别，用来抑制一些日志。
func SetLevel(level uint32) {
	atomic.StoreUint32(&logLevel, level)
}

// WithCooldownMillis 自定义记录堆栈的写入时间间隔。
func WithCooldownMillis(millis int) LogOption {
	return func(opts *logOptions) {
		opts.logStackCooldownMillis = millis
	}
}

// WithGzip 自动压缩日志文件。
func WithGzip() LogOption {
	return func(opts *logOptions) {
		opts.gzipEnabled = true
	}
}

// WithKeepDays 自定义日志文件保留天数。
func WithKeepDays(days int) LogOption {
	return func(opts *logOptions) {
		opts.keepDays = days
	}
}

// WithMaxBackups 自定义日志文件的最大保留份数。
func WithMaxBackups(count int) LogOption {
	return func(opts *logOptions) {
		opts.maxBackups = count
	}
}

// WithMaxSize 自定义日志文件最大尺寸（MB）
func WithMaxSize(size int) LogOption {
	return func(opts *logOptions) {
		opts.maxSize = size
	}
}

// WithRotation 自定义日志轮换规则。
func WithRotation(rule string) LogOption {
	return func(opts *logOptions) {
		opts.rotationRule = rule
	}
}

// Field 将 key, value 转换为 LogField。
func Field(key string, value any) LogField {
	switch val := value.(type) {
	case error:
		return LogField{Key: key, Value: val.Error()}
	case []error:
		var vs []string
		for _, err := range val {
			vs = append(vs, err.Error())
		}
		return LogField{Key: key, Value: vs}
	case time.Duration:
		return LogField{Key: key, Value: fmt.Sprint(val)}
	case []time.Duration:
		var durs []string
		for _, dur := range val {
			durs = append(durs, fmt.Sprint(dur))
		}
		return LogField{Key: key, Value: durs}
	case []time.Time:
		var ts []string
		for _, t := range val {
			ts = append(ts, fmt.Sprint(t))
		}
		return LogField{Key: key, Value: ts}
	case fmt.Stringer:
		return LogField{Key: key, Value: val.String()}
	case []fmt.Stringer:
		var vs []string
		for _, v := range val {
			vs = append(vs, v.String())
		}
		return LogField{Key: key, Value: vs}
	default:
		return LogField{Key: key, Value: val}
	}
}

// Info 将 v 写入访问日志。
func Info(v ...any) {
	writeInfo(fmt.Sprint(v...))
}

// Infof 将带有格式的 v 写入访问日志。
func Infof(format string, v ...any) {
	writeInfo(fmt.Sprintf(format, v...))
}

// Infov 使用 json 内容将 v 写入访问日志。
func Infov(v any) {
	writeInfo(v)
}

// Infow 将 msg 与字段一起写入访问日志。
func Infow(msg string, fields ...LogField) {
	writeInfo(msg, fields...)
}

// Alert 警报级别的 v 信息，并写入错误日志。
func Alert(v string) {
	getWriter().Alert(v)
}

// Debug writes v into access log.
func Debug(v ...any) {
	writeDebug(fmt.Sprint(v...))
}

// Debugf writes v with format into access log.
func Debugf(format string, v ...any) {
	writeDebug(fmt.Sprintf(format, v...))
}

// Debugv writes v into access log with json content.
func Debugv(v any) {
	writeDebug(v)
}

// Debugw writes msg along with fields into access log.
func Debugw(msg string, fields ...LogField) {
	writeDebug(msg, fields...)
}

// Error 将 v 写入错误日志。
func Error(v ...any) {
	writeError(fmt.Sprint(v...))
}

// Errorf 将带有格式的 v 写入错误日志。
func Errorf(format string, v ...any) {
	writeError(fmt.Errorf(format, v...).Error())
}

// ErrorStack 将 v 和调用堆栈一起写入错误日志。
func ErrorStack(v ...any) {
	writeStack(fmt.Sprint(v...))
}

// ErrorStackf 将格式化后的 v 和调用堆栈一起写入错误日志。
func ErrorStackf(format string, v ...any) {
	writeStack(fmt.Sprintf(format, v...))
}

// Errorv 将 json 编组后的 v 写入错误日志。
// 不带有调用堆栈，因为打包堆栈并不优雅。
func Errorv(v any) {
	writeError(v)
}

// Errorw 将 msg 与字段一起写入错误日志。
func Errorw(msg string, fields ...LogField) {
	writeError(msg, fields...)
}

// Severe 将 v 和调用堆栈一起写入严重错误日志。
func Severe(v ...any) {
	writeSevere(fmt.Sprint(v...))
}

// Severef 将带有格式的 v 和调用堆栈写入严重错误日志。
func Severef(format string, v ...any) {
	writeSevere(fmt.Sprintf(format, v...))
}

// Slow 将 v 写入慢执行日志。
func Slow(v ...any) {
	writeSlow(fmt.Sprint(v...))
}

// Slowf 将格式化后的 v 写入慢执行日志。
func Slowf(format string, v ...any) {
	writeSlow(fmt.Sprintf(format, v...))
}

// Slowv 将 json 编组后的 v 写入慢执行日志。
func Slowv(v any) {
	writeSlow(v)
}

// Sloww 将 msg 与字段一起写入慢执行日志。
func Sloww(msg string, fields ...LogField) {
	writeSlow(msg, fields...)
}

// Stat 将 v 写入统计日志。
func Stat(v ...any) {
	writeStat(fmt.Sprint(v...))
}

// Statf 将格式化后的 v 写入统计日志。
func Statf(format string, v ...any) {
	writeStat(fmt.Sprintf(format, v...))
}

func setupLogLevel(c Config) {
	switch c.Level {
	case levelInfo:
		SetLevel(InfoLevel)
	case levelError:
		SetLevel(ErrorLevel)
	case levelSevere:
		SetLevel(SevereLevel)
	}
}

func addCaller(fields ...LogField) []LogField {
	return append(fields, Field(callerKey, getCaller(callerDepth)))
}

func writeStat(val string) {
	if shallLogStat() && shallLog(InfoLevel) {
		getWriter().Stat(val, addCaller()...)
	}
}

func writeSlow(val any, fields ...LogField) {
	if shallLog(ErrorLevel) {
		getWriter().Slow(val, addCaller(fields...)...)
	}
}

func writeSevere(val string) {
	if shallLog(SevereLevel) {
		getWriter().Severe(fmt.Sprintf("%s\n%s", val, string(debug.Stack())))
	}
}

func writeError(val any, fields ...LogField) {
	if shallLog(ErrorLevel) {
		getWriter().Error(val, addCaller(fields...)...)
	}
}

func writeStack(val string) {
	if shallLog(ErrorLevel) {
		bs := debug.Stack()
		msg := fmt.Sprintf("%s\n%s", val, string(bs))
		getWriter().Stack(msg)
	}
}

func writeInfo(val any, fields ...LogField) {
	if shallLog(InfoLevel) {
		getWriter().Info(val, addCaller(fields...)...)
	}
}

func shallLog(level uint32) bool {
	return atomic.LoadUint32(&logLevel) <= level
}

func shallLogStat() bool {
	return atomic.LoadUint32(&disableStat) == 0
}

func writeDebug(val any, fields ...LogField) {
	if shallLog(DebugLevel) {
		getWriter().Debug(val, addCaller(fields...)...)
	}
}

func getWriter() Writer {
	w := writer.Load()
	if w == nil {
		w = writer.StoreIfNil(newConsoleWriter())
	}

	return w
}

func handleOptions(opts []LogOption) {
	for _, opt := range opts {
		opt(&options)
	}
}

func createOutput(path string) (io.WriteCloser, error) {
	if len(path) == 0 {
		return nil, ErrLogPathNotSet
	}

	switch options.rotationRule {
	case sizeRotationRule:
		return NewLogger(path, NewSizeLimitRotateRule(path, backupFileDelimiter, options.keepDays,
			options.maxSize, options.maxBackups, options.gzipEnabled), options.gzipEnabled)
	default:
		return NewLogger(path, DefaultRotateRule(path, backupFileDelimiter, options.keepDays,
			options.gzipEnabled), options.gzipEnabled)
	}
}

func setupWithFiles(c Config) error {
	w, err := newFileWriter(c)
	if err != nil {
		return err
	}

	SetWriter(w)
	return nil
}

func setupWithVolume(c Config) error {
	if len(c.ServiceName) == 0 {
		return ErrLogServiceNameNotSet
	}

	c.Path = path.Join(c.Path, c.ServiceName, sysx.Hostname())
	return setupWithFiles(c)
}

func setupWithConsole() {
	SetWriter(newConsoleWriter())
}
