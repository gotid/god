package conf

import (
	"fmt"
	"github.com/gotid/god/lib/iox"
	"os"
	"strconv"
	"strings"
	"sync"
)

// PropertyError 表示一个配置错误消息。
type PropertyError struct {
	error
	message string
}

// Properties 接口访问配置项的方法。
type Properties interface {
	GetString(key string) string
	SetString(key, value string)
	GetInt(key string) int
	SetInt(key string, value int)
	String() string
}

// 基于kv字典的配置项。
type mapBasedProperties struct {
	properties map[string]string
	lock       sync.RWMutex
}

// NewProperties 返回一个新的配置项 Properties。
func NewProperties() Properties {
	return &mapBasedProperties{
		properties: make(map[string]string),
	}
}

// LoadProperties 加载文件中的配置到配置项实例。
func LoadProperties(filename string, opts ...Option) (Properties, error) {
	lines, err := iox.ReadTextLines(filename, iox.WithoutBlank(), iox.OmitWithPrefix("#"))
	if err != nil {
		return nil, err
	}

	var opt options
	for _, o := range opts {
		o(&opt)
	}

	raw := make(map[string]string)
	for i := range lines {
		pair := strings.Split(lines[i], "=")
		if len(pair) != 2 {
			// 无效属性格式
			return nil, &PropertyError{
				error:   nil,
				message: fmt.Sprintf("无效的属性格式：%s", pair),
			}
		}

		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		if opt.env {
			raw[key] = os.ExpandEnv(value)
		} else {
			raw[key] = value
		}
	}

	return &mapBasedProperties{properties: raw}, nil
}

func (p *mapBasedProperties) GetString(key string) string {
	p.lock.RLock()
	ret := p.properties[key]
	p.lock.RUnlock()

	return ret
}

func (p *mapBasedProperties) SetString(key, value string) {
	p.lock.Lock()
	p.properties[key] = value
	p.lock.Unlock()
}

func (p *mapBasedProperties) GetInt(key string) int {
	p.lock.RLock()
	value, _ := strconv.Atoi(p.properties[key])
	p.lock.RUnlock()

	return value
}

func (p *mapBasedProperties) SetInt(key string, value int) {
	p.lock.Lock()
	p.properties[key] = strconv.Itoa(value)
	p.lock.Unlock()
}

func (p *mapBasedProperties) String() string {
	p.lock.RLock()
	ret := fmt.Sprintf("%s", p.properties)
	p.lock.RUnlock()

	return ret
}

func (e *PropertyError) Error() string {
	return e.message
}
