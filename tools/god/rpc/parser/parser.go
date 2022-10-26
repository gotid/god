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

// CamelCase 拷贝自 protobuf，细节详见 github.com/golang/protobuf@v1.4.2/protoc-gen-go/generator/generator.go:2648
func CamelCase(s string) string {
	if s == "" {
		return ""
	}

	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// 需要一个大写字母，丢弃 '_'。
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
