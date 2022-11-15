package filex

import "gopkg.in/cheggaaa/pb.v1"

type (
	// Scanner 接口用于按行读取。
	Scanner interface {
		// Scan 检查是否有可读内容。
		Scan() bool
		// Text 返回下一行文本字符串。
		Text() string
	}

	progressScanner struct {
		Scanner
		bar *pb.ProgressBar
	}
)

// NewProgressScanner 返回一个用于进度条指示器的 Scanner。
func NewProgressScanner(scanner Scanner, bar *pb.ProgressBar) Scanner {
	return &progressScanner{
		Scanner: scanner,
		bar:     bar,
	}
}

func (ps *progressScanner) Text() string {
	text := ps.Scanner.Text()
	ps.bar.Add64(int64(len(text)) + 1) // 考虑换行符

	return text
}
