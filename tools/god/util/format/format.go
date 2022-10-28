package format

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	flagGo       = "GO"
	flagDesigner = "DESIGNER"

	unknown style = iota
	title
	lower
	upper
)

type (
	style int

	styleFormat struct {
		before        string
		through       string
		after         string
		goStyle       style
		designerStyle style
	}
)

var ErrNamingFormat = errors.New("不支持的命名样式")

// FileNamingFormat 返回格式化后的文件名。
// 可通过 go 和 designer 这两个格式化字符串来定义命名样式，如蛇式：go_design，驼峰式: goDesign。
// 理论上甚至可以使用分隔符如 go#Designer，但还是要遵循操作系统的文件命名规范。
// 注意：FileNamingFormat 基于蛇式或驼峰。
func FileNamingFormat(format, content string) (string, error) {
	upperFormat := strings.ToUpper(format)
	indexGo := strings.Index(upperFormat, flagGo)
	indexDesigner := strings.Index(upperFormat, flagDesigner)
	if indexGo < 0 || indexDesigner < 0 || indexGo > indexDesigner {
		return "", ErrNamingFormat
	}

	var (
		before, through, after string
		flagGo, flagDesigner   string
		goStyle, designerStyle style
		err                    error
	)

	before = format[:indexGo]
	flagGo = format[indexGo : indexGo+2]
	through = format[indexGo+2 : indexDesigner]
	flagDesigner = format[indexDesigner : indexDesigner+8]
	after = format[indexDesigner+8:]

	goStyle, err = getStyle(flagGo)
	if err != nil {
		return "", err
	}
	designerStyle, err = getStyle(flagDesigner)
	if err != nil {
		return "", err
	}

	var formatStyle styleFormat
	formatStyle.goStyle = goStyle
	formatStyle.designerStyle = designerStyle
	formatStyle.before = before
	formatStyle.through = through
	formatStyle.after = after
	return doFormat(formatStyle, content)
}

func doFormat(format styleFormat, content string) (string, error) {
	fields, err := split(content)
	if err != nil {
		return "", err
	}

	var join []string
	for i, v := range fields {
		if i == 0 {
			join = append(join, transferTo(v, format.goStyle))
			continue
		}
		join = append(join, transferTo(v, format.designerStyle))
	}
	joined := strings.Join(join, format.through)

	return format.before + joined + format.after, nil
}

func transferTo(v string, style style) string {
	switch style {
	case upper:
		return strings.ToUpper(v)
	case lower:
		return strings.ToLower(v)
	case title:
		return strings.Title(v)
	default:
		return v
	}
}

func split(content string) ([]string, error) {
	var (
		list   []string
		reader = strings.NewReader(content)
		buffer = bytes.NewBuffer(nil)
	)

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if buffer.Len() > 0 {
					list = append(list, buffer.String())
				}
				return list, nil
			}
			return nil, err
		}

		if r == '_' {
			if buffer.Len() > 0 {
				list = append(list, buffer.String())
			}
			buffer.Reset()
			continue
		}

		if r >= 'A' && r <= 'Z' {
			if buffer.Len() > 0 {
				list = append(list, buffer.String())
			}
			buffer.Reset()
		}
		buffer.WriteRune(r)
	}
}

func getStyle(flag string) (style, error) {
	compare := strings.ToLower(flag)
	switch flag {
	case strings.ToLower(compare):
		return lower, nil
	case strings.ToUpper(compare):
		return upper, nil
	case strings.Title(compare):
		return title, nil
	default:
		return unknown, fmt.Errorf("意外的格式：%s", flag)
	}
}
