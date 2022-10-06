package iox

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

type (
	textReadOptions struct {
		keepSpace     bool
		withoutBlanks bool
		omitPrefix    string
	}

	// TextReadOption 自定义文本读取函数的方法。
	TextReadOption func(*textReadOptions)
)

// DupReadCloser 返回两个 io.ReadCloser，其中第一个将写入第二个。
func DupReadCloser(reader io.ReadCloser) (io.ReadCloser, io.ReadCloser) {
	var buf bytes.Buffer
	tee := io.TeeReader(reader, &buf)
	return io.NopCloser(tee), io.NopCloser(&buf)
}

// ReadBytes 精确读取长度为 len(buf) 的字节。
func ReadBytes(reader io.Reader, buf []byte) error {
	var got int

	for got < len(buf) {
		n, err := reader.Read(buf[got:])
		if err != nil {
			return err
		}

		got += n
	}

	return nil
}

// ReadText 读取 filename 文件的内容并去除两端空格。
func ReadText(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}

// ReadTextLines 按照读取选项打开并返回文本行切片。
func ReadTextLines(filename string, opts ...TextReadOption) ([]string, error) {
	var readOpts textReadOptions
	for _, opt := range opts {
		opt(&readOpts)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !readOpts.keepSpace {
			line = strings.TrimSpace(line)
		}
		if readOpts.withoutBlanks && len(line) == 0 {
			continue
		}
		if len(readOpts.omitPrefix) > 0 && strings.HasPrefix(line, readOpts.omitPrefix) {
			continue
		}

		lines = append(lines, line)
	}

	return lines, scanner.Err()
}

// KeepSpace 保留首尾空白的自定义读取函数以。
func KeepSpace() TextReadOption {
	return func(o *textReadOptions) {
		o.keepSpace = true
	}
}

// WithoutBlank 忽略空白行的自定义读取函数。
func WithoutBlank() TextReadOption {
	return func(o *textReadOptions) {
		o.withoutBlanks = true
	}
}

// OmitWithPrefix 忽略给定前缀的文本行的自定义读取函数。
func OmitWithPrefix(prefix string) TextReadOption {
	return func(o *textReadOptions) {
		o.omitPrefix = prefix
	}
}
