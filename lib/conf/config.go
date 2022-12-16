package conf

import (
	"fmt"
	"github.com/gotid/god/internal/encoding"
	"github.com/gotid/god/lib/jsonx"
	"github.com/gotid/god/lib/mapping"
	"log"
	"os"
	"path"
	"strings"
)

// 大小写字母的数值差值
const distanceBetweenUpperAndLower = 32

var loaders = map[string]func([]byte, any) error{
	".json": LoadFromJsonBytes,
	".yaml": LoadFromYamlBytes,
	".yml":  LoadFromYamlBytes,
}

// MustLoad 加载给定配置文件 path 至 v，遇错退出。
// 支持自定义选项，如使用环境变量。
func MustLoad(path string, v any, opts ...Option) {
	if err := Load(path, v, opts...); err != nil {
		log.Fatalf("错误：配置文件 %s，%s", path, err.Error())
	}
}

// Load 加载给定配置文件 file 至 v，支持 json|yaml 文件。
func Load(file string, v any, opts ...Option) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	loader, ok := loaders[strings.ToLower(path.Ext(file))]
	if !ok {
		return fmt.Errorf("配置文件类型仅支持 json|yaml|yml，错误文件格式：%s", file)
	}

	var opt options
	for _, o := range opts {
		o(&opt)
	}

	// 使用环境变量的情况
	if opt.env {
		return loader([]byte(os.ExpandEnv(string(content))), v)
	}

	return loader(content, v)
}

// LoadFromJsonBytes 从 json 字节切片中加载配置到变量 v。
func LoadFromJsonBytes(content []byte, v any) error {
	var m map[string]any
	if err := jsonx.Unmarshal(content, &m); err != nil {
		return err
	}

	return mapping.UnmarshalJsonMap(toCamelCaseKeyMap(m), v, mapping.WithCanonicalKeyFunc(toCamelCase))
}

// LoadFromYamlBytes 从 yaml 字节切片中加载配置到变量 v。
func LoadFromYamlBytes(content []byte, v any) error {
	bs, err := encoding.YamlToJson(content)
	if err != nil {
		return err
	}

	return LoadFromJsonBytes(bs, v)
}

func toCamelCase(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	var capNext bool
	boundary := true
	for _, v := range s {
		isCap := v >= 'A' && v <= 'Z'
		isLow := v >= 'a' && v <= 'z'
		if boundary && (isCap || isLow) {
			if capNext {
				if isLow {
					v -= distanceBetweenUpperAndLower
				}
			} else {
				if isCap {
					v += distanceBetweenUpperAndLower
				}
			}
			boundary = false
		}

		if isCap || isLow {
			buf.WriteRune(v)
			capNext = false
		} else if v == ' ' || v == '\t' {
			buf.WriteRune(v)
			capNext = false
			boundary = true
		} else if v == '_' {
			capNext = true
			boundary = true
		} else {
			buf.WriteRune(v)
			capNext = true
		}
	}

	return buf.String()
}

func toCamelCaseKeyMap(m map[string]any) map[string]any {
	ret := make(map[string]any)
	for k, v := range m {
		ret[toCamelCase(k)] = toCamelCaseInterface(v)
	}

	return ret
}

func toCamelCaseInterface(v any) any {
	switch vv := v.(type) {
	case map[string]any:
		return toCamelCaseKeyMap(vv)
	case []any:
		var arr []any
		for _, vvv := range vv {
			arr = append(arr, toCamelCaseInterface(vvv))
		}
		return arr
	default:
		return v
	}
}
