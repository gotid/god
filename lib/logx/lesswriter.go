package logx

import "io"

// 指定时间内仅写一次的日志写入器
type lessWriter struct {
	*limitedLogger
	writer io.Writer
}

// 新建一个指定时间内只记录一次的日志记录器
func newLessWriter(writer io.Writer, milliseconds int) *lessWriter {
	return &lessWriter{
		limitedLogger: newLimitedLogger(milliseconds),
		writer:        writer,
	}
}

func (w *lessWriter) Write(p []byte) (n int, err error) {
	w.logOrDiscard(func() {
		w.writer.Write(p)
	})
	return len(p), nil
}
