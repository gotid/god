package conf

import (
	"fmt"
	"github.com/gotid/god/lib/mapping"
	"log"
	"os"
	"path"
	"strings"
)

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
		return fmt.Errorf("配置文件类型仅支持 json|yaml|yml，错误：%s", file)
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
	return mapping.UnmarshalJsonBytes(content, v)
}

// LoadFromYamlBytes 从 yaml 字节切片中加载配置到变量 v。
func LoadFromYamlBytes(content []byte, v any) error {
	return mapping.UnmarshalYamlBytes(content, v)
}
