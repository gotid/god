package sortedmap

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/gotid/god/tools/god/util/stringx"
	"strings"
)

var (
	ErrInvalidKVExpression = errors.New("无效的 key-value 表达式")
	ErrInvalidKVS          = errors.New("kv 长度必须是偶数")
)

// SortedMap 是一个排序后的字典。
type SortedMap struct {
	kv   *list.List
	keys map[interface{}]*list.Element
}

type KV []interface{}

func (m *SortedMap) Format() []string {
	format := make([]string, 0)
	m.Range(func(key, val interface{}) {
		format = append(format, fmt.Sprintf("%s=%s", key, val))
	})
	return format
}

// Range 使用给定的迭代函数，遍历有序字典中的每个元素。
func (m *SortedMap) Range(iterator func(key interface{}, val interface{})) {
	next := m.kv.Front()
	for next != nil {
		value := next.Value.(KV)
		iterator(value[0], value[1])
		next = next.Next()
	}
}

func (m *SortedMap) SetKV(key, val interface{}) {
	e, ok := m.keys[key]
	if !ok {
		e = m.kv.PushBack(KV{key, val})
	} else {
		e.Value.(KV)[1] = val
	}
	m.keys[key] = e
}

// SetExpression 按照 key-value 结构的表达式，设置一对字典成员。
func (m *SortedMap) SetExpression(expression string) (key, value interface{}, err error) {
	idx := strings.Index(expression, "=")
	if idx == -1 {
		return "", "", ErrInvalidKVExpression
	}

	key = expression[:idx]
	if len(expression) == idx {
		value = ""
	} else {
		value = expression[idx+1:]
	}

	if keys, ok := key.(string); ok && stringx.ContainsWhitespace(keys) {
		return "", "", ErrInvalidKVExpression
	}
	if values, ok := value.(string); ok && stringx.ContainsWhitespace(values) {
		return "", "", ErrInvalidKVExpression
	}
	if len(key.(string)) == 0 {
		return "", "", ErrInvalidKVExpression
	}

	m.SetKV(key, value)

	return
}

// GetString 获取有序字典中给定键的字符串值。
func (m *SortedMap) GetString(key interface{}) (string, bool) {
	value, ok := m.Get(key)
	if !ok {
		return "", false
	}
	vs, ok := value.(string)
	return vs, ok
}

// Get 获取有序字典中给定键的值。
func (m *SortedMap) Get(key interface{}) (interface{}, bool) {
	e, ok := m.keys[key]
	if !ok {
		return nil, false
	}

	return e.Value.(KV)[1], true
}

// GetOr 获取有序字典中给定键的值，若不存在则返回给定的默认值。
func (m *SortedMap) GetOr(key, dft interface{}) interface{} {
	e, ok := m.keys[key]
	if !ok {
		return dft
	}

	return e.Value.(KV)[1]
}

// GetStringOr 获取有序字典给定键的字符串值，若不存在则返回给定的默认字符串。
func (m *SortedMap) GetStringOr(key, dft string) string {
	value, ok := m.GetString(key)
	if !ok {
		return dft
	}

	return value
}

// HasKey 判断字典中是否有给定的键。
func (m *SortedMap) HasKey(key interface{}) bool {
	_, ok := m.keys[key]
	return ok
}

// New 返回一个 SortedMap。
func New() *SortedMap {
	return &SortedMap{
		kv:   list.New(),
		keys: make(map[interface{}]*list.Element),
	}
}
