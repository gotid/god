package logx

import "io"

type lessWriter struct {
	*limitedExecutor
	writer io.Writer
}

// 返回一个间隔时间大于给定毫秒数的受限日志编写器。
func newLessWriter(writer io.Writer, milliseconds int) *lessWriter {
	return &lessWriter{
		limitedExecutor: newLimitedExecutor(milliseconds),
		writer:          writer,
	}
}

func (w *lessWriter) Write(p []byte) (int, error) {
	w.logOrDiscard(func() {
		w.writer.Write(p)
	})
	return len(p), nil
}
