package parser

import (
	"github.com/emicklei/proto"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DefaultProtoParser 定义了一个空白结构体。
type DefaultProtoParser struct{}

// NewDefaultProtoParser 返回一个 DefaultProtoParser 实例。
func NewDefaultProtoParser() *DefaultProtoParser {
	return &DefaultProtoParser{}
}

// Parse 将给定的 proto 文件解析至 golang 结构体。
// 便于后续 rpc 服务生成和使用。
func (p *DefaultProtoParser) Parse(src string, multiple ...bool) (Proto, error) {
	var ret Proto

	abs, err := filepath.Abs(src)
	if err != nil {
		return Proto{}, err
	}

	r, err := os.Open(abs)
	if err != nil {
		return ret, err
	}
	defer r.Close()

	parser := proto.NewParser(r)
	set, err := parser.Parse()
	if err != nil {
		return ret, err
	}

	var serviceList Services
	proto.Walk(
		set,
		proto.WithImport(func(i *proto.Import) {
			ret.Import = append(ret.Import, Import{Import: i})
		}),
		proto.WithMessage(func(message *proto.Message) {
			ret.Message = append(ret.Message, Message{Message: message})
		}),
		proto.WithPackage(func(p *proto.Package) {
			ret.Package = Package{Package: p}
		}),
		proto.WithService(func(service *proto.Service) {
			svc := Service{Service: service}
			elements := service.Elements
			for _, el := range elements {
				v, _ := el.(*proto.RPC)
				if v == nil {
					continue
				}
				svc.RPC = append(svc.RPC, &RPC{RPC: v})
			}

			serviceList = append(serviceList, svc)
		}),
		proto.WithOption(func(option *proto.Option) {
			if option.Name == "go_package" {
				ret.GoPackage = option.Constant.Source
			}
		}),
	)
	if err = serviceList.validate(abs, multiple...); err != nil {
		return ret, err
	}

	if len(ret.GoPackage) == 0 {
		ret.GoPackage = ret.Package.Name
	}

	ret.PbPackage = GoSanitized(filepath.Base(ret.GoPackage))
	ret.Src = src
	ret.Name = filepath.Base(abs)
	ret.Service = serviceList

	return ret, nil
}

// GoSanitized 拷贝自 protobuf，详见 google.golang.org/protobuf@v1.25.0/internal/strs/strings.go:71。
func GoSanitized(s string) string {
	// 整理输入字符串为有效字符（仅保留字母和数字，其他转为下划线）
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}

		return '_'
	}, s)

	// 若与 Go 关键字冲突或标识符不是字母，则增加下划线作为前缀。
	r, _ := utf8.DecodeRuneInString(s)
	if token.Lookup(s).IsKeyword() || !unicode.IsLetter(r) {
		return "_" + s
	}

	return s
}
