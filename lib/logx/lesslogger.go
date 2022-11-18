package logx

// LessLogger 是一个记录器，它控制在给定的持续时间内记录一次。
type LessLogger struct {
	*limitedExecutor
}

// NewLessLogger returns a LessLogger.
func NewLessLogger(milliseconds int) *LessLogger {
	return &LessLogger{
		limitedExecutor: newLimitedExecutor(milliseconds),
	}
}

// Error 将 v 记录到错误日志中，如果在给定的持续时间内不止一次，则将其丢弃。
func (logger *LessLogger) Error(v ...any) {
	logger.logOrDiscard(func() {
		Error(v...)
	})
}

// Errorf 将带有格式的 v 记录到错误日志中，如果在给定的持续时间内不止一次，则将其丢弃。
func (logger *LessLogger) Errorf(format string, v ...any) {
	logger.logOrDiscard(func() {
		Errorf(format, v...)
	})
}
