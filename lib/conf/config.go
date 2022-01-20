package conf

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/gotid/god/lib/mapping"
)

var loaders = map[string]func([]byte, interface{}) error{
	".json": LoadConfigFromJsonBytes,
	".yaml": LoadConfigFromYamlBytes,
	".yml":  LoadConfigFromYamlBytes,
}

func LoadConfigFromJsonBytes(content []byte, v interface{}) error {
	return mapping.UnmarshalJsonBytes(content, v)
}

func LoadConfigFromYamlBytes(bytes []byte, v interface{}) error {
	return mapping.UnmarshalYamlBytes(bytes, v)
}

func LoadConfig(filename string, v interface{}, opts ...Option) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	loader, ok := loaders[path.Ext(filename)]
	if !ok {
		return fmt.Errorf("不识别的配置文件类型：%s", filename)
	}

	var opt options
	for _, o := range opts {
		o(&opt)
	}

	if opt.env {
		return loader([]byte(os.ExpandEnv(string(content))), v)
	} else {
		return loader(content, v)
	}
}

func MustLoad(path string, v interface{}, opts ...Option) {
	if err := LoadConfig(path, v, opts...); err != nil {
		log.Fatalf("错误：配置文件 - %s, %s", path, err.Error())
	}
}
