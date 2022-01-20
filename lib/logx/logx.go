package logx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"git.zc0901.com/go/god/lib/iox"
	"git.zc0901.com/go/god/lib/sysx"
	"git.zc0901.com/go/god/lib/timex"
)

// 日志级别值
const (
	// InfoLevel 记录所有日志。
	InfoLevel = iota
	// DebugLevel 调试级别
	DebugLevel
	// ErrorLevel 包括错误日志、慢日志和堆栈日志。
	ErrorLevel
	// SevereLevel 仅记录严重错误。
	SevereLevel
)

const (
	accessFilename = "access.log"
	errorFilename  = "error.log"
	severeFilename = "severe.log"
	slowFilename   = "slow.log"
	statFilename   = "stat.log"

	consoleMode = "console" // 命令行模式
	volumeMode  = "volume"  // k8s 模式

	alertLevel  = "alert"  // 警告级
	infoLevel   = "info"   // 信息级
	debugLevel  = "debug"  // 调试级
	errorLevel  = "error"  // 错误级
	serverLevel = "severe" // 严重级
	fatalLevel  = "fatal"  // 致命级
	slowLevel   = "slow"   // 慢级别
	statLevel   = "stat"   // 统计级

	callerInnerDepth    = 5 // 堆栈调用深度
	flags               = 0x0
	backupFileDelimiter = "-" // 备份文件分隔符
)

var (
	infoLog   io.WriteCloser // 信息日志
	debugLog  io.WriteCloser // 调试日志
	errorLog  io.WriteCloser // 错误日志
	severeLog io.WriteCloser // 严重日志
	slowLog   io.WriteCloser // 慢日志
	statLog   io.WriteCloser // 统计日志
	stackLog  io.Writer      // 堆栈日志

	initialized  uint32    // 初始状态
	logLevel     uint32    // 日志级别
	disableStat  uint32    // 禁用统计
	writeConsole bool      // 写控制台
	once         sync.Once // 一次操作对象
	options      logOptions

	timeFormat = "2006-01-02T15:04:05.000Z07" // 日期格式

	// ErrLogPathNotSet 指示日志路径未设置的错误。
	ErrLogPathNotSet = errors.New("日志路径必须设置")
	// ErrLogNotInitialized 指示日志尚未初始化的错误。
	ErrLogNotInitialized = errors.New("日志尚未初始化")
	// ErrLogServiceNameNotSet 指示日志名称未设置的错误。
	ErrLogServiceNameNotSet = errors.New("日志服务名称必须设置")
)

type (
	// 日志结构
	logEntry struct {
		Timestamp string      `json:"@timestamp"`
		Level     string      `json:"level"`
		Duration  string      `json:"duration,omitempty"`
		Content   interface{} `json:"content"`
	}

	// 日志配置选项
	logOptions struct {
		gzipEnabled           bool
		logStackCooldownMills int
		keepDays              int
	}

	LogOption func(options *logOptions)

	// Logger 用于 durationLogger/traceLogger
	Logger interface {
		Info(...interface{})
		Infof(string, ...interface{})
		Infov(interface{})
		Debug(...interface{})
		Debugf(string, ...interface{})
		Debugv(interface{})
		Error(...interface{})
		Errorf(string, ...interface{})
		Errorv(interface{})
		Slow(...interface{})
		Slowf(string, ...interface{})
		Slowv(interface{})
		WithDuration(time.Duration) Logger
	}
)

// MustSetup 使用指定配置设置日志。出错退出。
func MustSetup(c LogConf) {
	Must(Setup(c))
}

// Must 检查错误是否为空，否则记录错误并退出。
func Must(err error) {
	if err != nil {
		msg := formatWithCaller(err.Error(), 3)
		log.Print(msg)
		outputText(severeLog, fatalLevel, msg)
		os.Exit(1)
	}
}

// Setup 设置 logx。如已设置则返回。
// 允许不同的服务框架调用多次 Setup。
func Setup(c LogConf) error {
	if len(c.TimeFormat) > 0 {
		timeFormat = c.TimeFormat
	}

	switch c.Mode {
	case consoleMode:
		setupWithConsole(c)
		return nil
	case volumeMode:
		return setupWithVolume(c)
	default:
		return setupWithFiles(c)
	}
}

// Close 关闭日志。
func Close() error {
	if writeConsole {
		return nil
	}

	if atomic.LoadUint32(&initialized) == 0 {
		return ErrLogNotInitialized
	}

	atomic.StoreUint32(&initialized, 0)

	loggers := []io.WriteCloser{infoLog, errorLog, severeLog, slowLog, statLog}
	for _, logger := range loggers {
		if logger != nil {
			if err := logger.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Disable 禁用日志。
func Disable() {
	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)

		infoLog = iox.NopCloser(ioutil.Discard)
		debugLog = iox.NopCloser(ioutil.Discard)
		errorLog = iox.NopCloser(ioutil.Discard)
		severeLog = iox.NopCloser(ioutil.Discard)
		slowLog = iox.NopCloser(ioutil.Discard)
		statLog = iox.NopCloser(ioutil.Discard)
		stackLog = ioutil.Discard
	})
}

// DisableStat 禁用统计日志。
func DisableStat() {
	atomic.StoreUint32(&disableStat, 1)
}

// SetLevel 设置日志级别。可被用于抑制某些日志。
func SetLevel(level uint32) {
	atomic.StoreUint32(&logLevel, level)
}

func WithKeepDays(days int) LogOption {
	return func(opts *logOptions) {
		opts.keepDays = days
	}
}

func WithGzip() LogOption {
	return func(opts *logOptions) {
		opts.gzipEnabled = true
	}
}

func WithCooldownMillis(millis int) LogOption {
	return func(opts *logOptions) {
		opts.logStackCooldownMills = millis
	}
}

// Alert 输出警报并写入错误日志。
func Alert(v string) {
	outputText(errorLog, alertLevel, v)
}

// Info 将值写入访问日志。
func Info(v ...interface{}) {
	syncInfoText(fmt.Sprint(v...))
}

// Infof 将格式化的值写入访问日志。
func Infof(format string, args ...interface{}) {
	syncInfoText(fmt.Sprintf(format, args...))
}

// Infov 将值以JSON格式写入访问日志。
func Infov(v interface{}) {
	syncInfoAny(v)
}

// Debug 将值写入调试日志。
func Debug(v ...interface{}) {
	syncDebugText(fmt.Sprint(v...))
}

// Debugf 将格式化的值写入调试日志。
func Debugf(format string, args ...interface{}) {
	syncDebugText(fmt.Sprintf(format, args...))
}

// Debugv 将值以JSON格式写入调试日志。
func Debugv(v interface{}) {
	syncDebugAny(v)
}

// Slow 将值写入慢日志。
func Slow(v ...interface{}) {
	syncSlowText(fmt.Sprint(v...))
}

// Slowf 将格式化值写入慢日志。
func Slowf(format string, v ...interface{}) {
	syncSlowText(fmt.Sprintf(format, v...))
}

// Slowv 将值以JSON格式写入慢日志。
func Slowv(v interface{}) {
	syncSlowAny(v)
}

// Error 将值写入错误日志。
func Error(v ...interface{}) {
	ErrorCaller(1, v...)
}

// Errorf 将带格式的值写入错误日志。
func Errorf(format string, v ...interface{}) {
	ErrorCallerf(1, format, v...)
}

// ErrorCaller 将带上下文的值写入错误日志。
func ErrorCaller(callDepth int, v ...interface{}) {
	syncErrorText(fmt.Sprint(v...), callDepth+callerInnerDepth)
}

// ErrorCallerf 将带上下文的格式化值写入错误日志。
func ErrorCallerf(callDepth int, format string, v ...interface{}) {
	syncErrorText(fmt.Errorf(format, v...).Error(), callDepth+callerInnerDepth)
}

// ErrorStack 将值和调用堆栈一起写入错误日志。
func ErrorStack(v ...interface{}) {
	syncStack(fmt.Sprint(v...))
}

// ErrorStackf 将格式化的值和调用堆栈一起写入错误日志。
func ErrorStackf(format string, v ...interface{}) {
	syncStack(fmt.Sprintf(format, v...))
}

// Errorv 将值以JSON格式写入错误日志。
// 因调用堆栈打包不优雅故未打包。
func Errorv(v interface{}) {
	syncErrorAny(v)
}

// Severe 将值写入严重日志。
func Severe(v ...interface{}) {
	syncSevere(fmt.Sprint(v...))
}

// Severef 将格式化值写入严重日志。
func Severef(format string, v ...interface{}) {
	syncSevere(fmt.Sprintf(format, v...))
}

func Stat(v ...interface{}) {
	syncStat(fmt.Sprint(v...))
}

func Statf(format string, v ...interface{}) {
	syncStat(fmt.Sprintf(format, v...))
}

func setupLogLevel(c LogConf) {
	switch c.Level {
	case infoLevel:
		SetLevel(InfoLevel)
	case errorLevel:
		SetLevel(ErrorLevel)
	case serverLevel:
		SetLevel(SevereLevel)
	}
}

func shouldLog(level uint32) bool {
	return atomic.LoadUint32(&logLevel) <= level
}

func shouldLogStat() bool {
	return atomic.LoadUint32(&disableStat) == 0
}

func setupWithConsole(c LogConf) {
	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)
		writeConsole = true
		setupLogLevel(c)

		infoLog = newLogWriter(log.New(os.Stdout, "", flags))
		errorLog = newLogWriter(log.New(os.Stderr, "", flags))
		severeLog = newLogWriter(log.New(os.Stderr, "", flags))
		slowLog = newLogWriter(log.New(os.Stderr, "", flags))
		stackLog = newLessWriter(errorLog, options.logStackCooldownMills)
		statLog = infoLog
	})
}

func setupWithFiles(c LogConf) error {
	var opts []LogOption
	var err error

	if len(c.Path) == 0 {
		return ErrLogPathNotSet
	}

	opts = append(opts, WithCooldownMillis(c.StackCooldownMillis))
	if c.Compress {
		opts = append(opts, WithGzip())
	}
	if c.KeepDays > 0 {
		opts = append(opts, WithKeepDays(c.KeepDays))
	}

	accessFile := path.Join(c.Path, accessFilename)
	errorFile := path.Join(c.Path, errorFilename)
	severeFile := path.Join(c.Path, severeFilename)
	slowFile := path.Join(c.Path, slowFilename)
	statFile := path.Join(c.Path, statFilename)

	once.Do(func() {
		atomic.StoreUint32(&initialized, 1)
		handleOptions(opts)
		setupLogLevel(c)

		if infoLog, err = createOutput(accessFile); err != nil {
			return
		}

		if errorLog, err = createOutput(errorFile); err != nil {
			return
		}

		if severeLog, err = createOutput(severeFile); err != nil {
			return
		}

		if slowLog, err = createOutput(slowFile); err != nil {
			return
		}

		if statLog, err = createOutput(statFile); err != nil {
			return
		}

		stackLog = newLessWriter(errorLog, options.logStackCooldownMills)
	})

	return err
}

func setupWithVolume(c LogConf) error {
	if len(c.ServiceName) == 0 {
		return ErrLogServiceNameNotSet
	}

	c.Path = path.Join(c.Path, c.ServiceName, sysx.Hostname())
	return setupWithFiles(c)
}

func handleOptions(opts []LogOption) {
	for _, opt := range opts {
		opt(&options)
	}
}

func createOutput(filename string) (io.WriteCloser, error) {
	if len(filename) == 0 {
		return nil, ErrLogPathNotSet
	}

	return NewLogger(filename,
		DefaultRotateRule(filename, backupFileDelimiter, options.keepDays, options.gzipEnabled),
		options.gzipEnabled)
}

func syncInfoText(msg string) {
	if shouldLog(InfoLevel) {
		outputText(infoLog, infoLevel, msg)
	}
}

func syncInfoAny(v interface{}) {
	if shouldLog(InfoLevel) {
		outputAny(infoLog, infoLevel, v)
	}
}

func syncDebugText(msg string) {
	if shouldLog(DebugLevel) {
		outputText(infoLog, debugLevel, msg)
	}
}

func syncDebugAny(v interface{}) {
	if shouldLog(DebugLevel) {
		outputAny(infoLog, debugLevel, v)
	}
}

func syncSlowAny(v interface{}) {
	if shouldLog(ErrorLevel) {
		outputAny(slowLog, slowLevel, v)
	}
}

func syncSlowText(msg string) {
	if shouldLog(ErrorLevel) {
		outputText(slowLog, slowLevel, msg)
	}
}

func syncErrorAny(v interface{}) {
	if shouldLog(ErrorLevel) {
		outputAny(errorLog, errorLevel, v)
	}
}

func syncErrorText(msg string, callDepth int) {
	if shouldLog(ErrorLevel) {
		outputError(errorLog, msg, callDepth)
	}
}

func syncSevere(msg string) {
	if shouldLog(SevereLevel) {
		outputText(severeLog, serverLevel, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
	}
}

func syncStack(msg string) {
	if shouldLog(ErrorLevel) {
		outputText(stackLog, errorLevel, fmt.Sprintf("%s\n%s", msg, string(debug.Stack())))
	}
}

func syncStat(msg string) {
	if shouldLogStat() && shouldLog(InfoLevel) {
		outputText(statLog, statLevel, msg)
	}
}

func outputError(writer io.WriteCloser, msg string, callDepth int) {
	content := formatWithCaller(msg, callDepth)
	outputText(writer, errorLevel, content)
}

func outputAny(writer io.Writer, level string, val interface{}) {
	outputJson(writer, logEntry{
		Timestamp: getTimestamp(),
		Level:     level,
		Content:   val,
	})
}

func outputText(writer io.Writer, level, msg string) {
	outputJson(writer, logEntry{
		Timestamp: getTimestamp(),
		Level:     level,
		Content:   msg,
	})
}

func outputJson(writer io.Writer, info interface{}) {
	if content, err := json.Marshal(info); err != nil {
		log.Println(err.Error())
	} else if atomic.LoadUint32(&initialized) == 0 || writer == nil {
		log.Println(string(content))
	} else {
		writer.Write(append(content, '\n'))
	}
}

func formatWithCaller(msg string, callDepth int) string {
	var b strings.Builder

	caller := getCaller(callDepth)
	if len(caller) > 0 {
		b.WriteString(caller)
		b.WriteByte(' ')
	}
	b.WriteString(msg)

	return b.String()
}

func getCaller(callDepth int) string {
	var b strings.Builder

	_, file, line, ok := runtime.Caller(callDepth)
	if ok {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		b.WriteString(short)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(line))
	}

	return b.String()
}

func getTimestamp() string {
	return timex.Time().Format(timeFormat)
}

// 日志输出器
type logWriter struct {
	logger *log.Logger
}

func newLogWriter(logger *log.Logger) logWriter {
	return logWriter{
		logger: logger,
	}
}

func (w logWriter) Close() error {
	return nil
}

func (w logWriter) Write(data []byte) (int, error) {
	w.logger.Print(string(data))
	return len(data), nil
}
