package conf

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gotid/god/lib/iox"
)

type PropertyError struct {
	error
	message string
}

type Properties interface {
	GetString(key string) string
	SetString(key, value string)
	GetInt(key string) int
	SetInt(key string, value int)
	ToString() string
}

type mapBasedProperties struct {
	properties map[string]string
	lock       sync.RWMutex
}

func LoadProperties(filename string) (Properties, error) {
	lines, err := iox.ReadTextLines(filename, iox.WithoutBlank(), iox.OmitWithPrefix("#"))
	if err != nil {
		return nil, err
	}

	raw := make(map[string]string)
	for i := range lines {
		pair := strings.Split(lines[i], "=")
		if len(pair) != 2 {
			return nil, &PropertyError{message: fmt.Sprintf("无效的属性格式：%s", pair)}
		}

		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])
		raw[key] = value
	}

	return &mapBasedProperties{
		properties: raw,
	}, nil
}

func (config *mapBasedProperties) GetString(key string) string {
	config.lock.RLock()
	ret := config.properties[key]
	config.lock.RUnlock()

	return ret
}

func (config *mapBasedProperties) SetString(key, value string) {
	config.lock.Lock()
	config.properties[key] = value
	config.lock.Unlock()
}

func (config *mapBasedProperties) GetInt(key string) int {
	config.lock.RLock()
	ret, _ := strconv.Atoi(config.properties[key])
	config.lock.RUnlock()

	return ret
}

func (config *mapBasedProperties) SetInt(key string, value int) {
	config.lock.Lock()
	config.properties[key] = strconv.Itoa(value)
	config.lock.Unlock()
}

func (config *mapBasedProperties) ToString() string {
	config.lock.RLock()
	ret := fmt.Sprintf("%s", config.properties)
	config.lock.RUnlock()

	return ret
}

func (configError *PropertyError) Error() string {
	return configError.message
}

func NewProperties() Properties {
	return &mapBasedProperties{
		properties: make(map[string]string),
	}
}
