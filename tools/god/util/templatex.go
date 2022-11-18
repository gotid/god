package util

import (
	"bytes"
	goformat "go/format"
	"os"
	"text/template"

	"github.com/gotid/god/tools/god/internal/errorx"
	"github.com/gotid/god/tools/god/util/pathx"
)

const regularPerm = 0o666

// DefaultTemplate 是一个提供文本模板 text/template 操作的工具。
type DefaultTemplate struct {
	name  string
	text  string
	goFmt bool
}

// With 返回一个给定名称的 DefaultTemplate 实例。
func With(name string) *DefaultTemplate {
	return &DefaultTemplate{name: name}
}

// Parse 将文本作为模板 DefaultTemplate 的文本。
func (t *DefaultTemplate) Parse(text string) *DefaultTemplate {
	t.text = text
	return t
}

// GoFmt 设置是否需要格式化生成的代码。
func (t *DefaultTemplate) GoFmt(format bool) *DefaultTemplate {
	t.goFmt = format
	return t
}

// SaveTo 写入代码到给定的目标路径。
func (t *DefaultTemplate) SaveTo(data any, path string, forceUpdate bool) error {
	if pathx.FileExists(path) && !forceUpdate {
		return nil
	}

	output, err := t.Execute(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, output.Bytes(), regularPerm)
}

// Execute 返回模板执行后的代码。
func (t *DefaultTemplate) Execute(data any) (*bytes.Buffer, error) {
	temp, err := template.New(t.name).Parse(t.text)
	if err != nil {
		return nil, errorx.Wrap(err, "模板解析错误：", t.text)
	}

	buf := new(bytes.Buffer)
	if err = temp.Execute(buf, data); err != nil {
		return nil, errorx.Wrap(err, "模板执行错误：", t.text)
	}

	if !t.goFmt {
		return buf, nil
	}

	formattedOutput, err := goformat.Source(buf.Bytes())
	if err != nil {
		return nil, errorx.Wrap(err, "go 格式化错误：", buf.String())
	}

	buf.Reset()
	buf.Write(formattedOutput)
	return buf, nil
}
