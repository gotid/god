package filex

import (
	"errors"
	"os"
)

var errExceedFileSize = errors.New("偏移量：错误或超过文件大小")

// RangeReader 用于从一个文件读取区间内容。
type RangeReader struct {
	file  *os.File
	start int64
	stop  int64
}

// NewRangeReader 返回一个 RangeReader，它将从一个文件读取给定区间的内容。
func NewRangeReader(file *os.File, start, stop int64) *RangeReader {
	return &RangeReader{
		file:  file,
		start: start,
		stop:  stop,
	}
}

// Read 读取给定区间的内容至字节数组 p。
func (r *RangeReader) Read(p []byte) (n int, err error) {
	stat, err := r.file.Stat()
	if err != nil {
		return 0, err
	}

	if r.stop < r.start || r.start >= stat.Size() {
		return 0, errExceedFileSize
	}

	if r.stop-r.start < int64(len(p)) {
		p = p[:r.stop-r.start]
	}

	n, err = r.file.ReadAt(p, r.start)
	if err != nil {
		return n, err
	}

	r.start += int64(n)

	return
}
