package logx

import "errors"

const (
	// InfoLevel 记录一切
	InfoLevel uint32 = iota
	// ErrorLevel 包括错误、慢执行和堆栈。
	ErrorLevel
	// SevereLevel 仅记录严重的信息。
	SevereLevel
)

const (
	jsonEncodingType = iota
	plainEncodingType

	plainEncoding    = "plain"
	plainEncodingSep = '\t'
	sizeRotationRule = "size"
)

const (
	accessFilename = "access.log"
	errorFilename  = "error.log"
	severeFilename = "severe.log"
	slowFilename   = "slow.log"
	statFilename   = "stat.log"

	fileMode   = "file"
	volumeMode = "volume"

	levelAlert  = "alert"
	levelInfo   = "info"
	levelError  = "error"
	levelSevere = "severe"
	levelFatal  = "fatal"
	levelSlow   = "slow"
	levelStat   = "stat"

	backupFileDelimiter = "-"
	flags               = 0x0
)

const (
	levelKey     = "level"
	callerKey    = "caller"
	contentKey   = "content"
	durationKey  = "duration"
	traceKey     = "trace"
	spanKey      = "span"
	timestampKey = "@timestamp"
)

var (
	// ErrLogPathNotSet 表示一个日志路径未设置的错误。
	ErrLogPathNotSet = errors.New("日志路径必须设置")
	// ErrLogServiceNameNotSet 表示一个日志服务名称未设置的错误。
	ErrLogServiceNameNotSet = errors.New("日志服务名必须设置")
)
