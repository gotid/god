package iox

import (
	"bufio"
	"io"
	"strings"
)

// TextLineScanner 是一个可以扫描给定阅读器文本行的扫描器。
type TextLineScanner struct {
	reader  *bufio.Reader
	hasNext bool
	line    string
	err     error
}

// NewTextLineScanner 返回一个给定阅读器的文本行扫描器 TextLineScanner。
func NewTextLineScanner(reader io.Reader) *TextLineScanner {
	return &TextLineScanner{
		reader:  bufio.NewReader(reader),
		hasNext: true,
	}
}

// Scan 判断扫描器是否有可扫描的文本行。
func (s *TextLineScanner) Scan() bool {
	if !s.hasNext {
		return false
	}

	line, err := s.reader.ReadString('\n')
	s.line = strings.TrimRight(line, "\n")
	if err == io.EOF {
		s.hasNext = false
		return true
	} else if err != nil {
		s.err = err
		return false
	}

	return true
}

// Line 返回文本的下一行。
func (s *TextLineScanner) Line() (string, error) {
	return s.line, s.err
}
