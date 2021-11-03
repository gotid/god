package internal

import (
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/syncx"
	"google.golang.org/grpc/grpclog"
)

const errorLevel = 2

type Logger struct{}

func InitLogger() {
	syncx.Once(func() {
		grpclog.SetLoggerV2(new(Logger))
	})
}

func (l *Logger) Info(args ...interface{}) {
	// 忽略内置 grpc 信息
}

func (l *Logger) Infoln(args ...interface{}) {
	// 忽略内置 grpc 信息
}

func (l *Logger) Infof(format string, args ...interface{}) {
	// 忽略内置 grpc 信息
}

func (l *Logger) Warning(args ...interface{}) {
	// 忽略内置 grpc 警告
}

func (l *Logger) Warningln(args ...interface{}) {
	// 忽略内置 grpc 警告
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	// 忽略内置 grpc 警告
}

func (l *Logger) Error(args ...interface{}) {
	logx.Error(args...)
}

func (l *Logger) Errorln(args ...interface{}) {
	logx.Error(args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	logx.Errorf(format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	logx.Error(args...)
}

func (l *Logger) Fatalln(args ...interface{}) {
	logx.Error(args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	logx.Errorf(format, args...)
}

func (l *Logger) V(level int) bool {
	return level >= errorLevel
}
