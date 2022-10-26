package collection

import (
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/logx"
)

const (
	unmanaged = iota
	untyped
	intType
	int64Type
	uintType
	uint64Type
	stringType
)

// Set 用于并发，非线程安全，确保在同步状态下使用。
type Set struct {
	data map[interface{}]lang.PlaceholderType
	tp   int
}

// NewSet 返回一个托管 Set，只能存放相同类型的值。
func NewSet() *Set {
	return &Set{
		data: make(map[interface{}]lang.PlaceholderType),
		tp:   untyped,
	}
}

// NewUnmanagedSet 返回一个非托管 Set，可以存放不同类型的值。
func NewUnmanagedSet() *Set {
	return &Set{
		data: make(map[interface{}]lang.PlaceholderType),
		tp:   unmanaged,
	}
}

// Add 添加 i 到集合 s。
func (s *Set) Add(i ...interface{}) {
	for _, each := range i {
		s.add(each)
	}
}

// AddInt 添加 int 值到集合 s。
func (s *Set) AddInt(ii ...int) {
	for _, each := range ii {
		s.add(each)
	}
}

// AddInt64 添加 int64 值到集合 s。
func (s *Set) AddInt64(ii ...int64) {
	for _, each := range ii {
		s.add(each)
	}
}

// AddUint 添加 uint 值到集合 s。
func (s *Set) AddUint(ii ...uint) {
	for _, each := range ii {
		s.add(each)
	}
}

// AddUint64 添加 uint64 值到集合 s。
func (s *Set) AddUint64(ii ...uint64) {
	for _, each := range ii {
		s.add(each)
	}
}

// AddStr 添加 string 值到集合 s。
func (s *Set) AddStr(ss ...string) {
	for _, each := range ss {
		s.add(each)
	}
}

// Contains 检查 i 是否存在于集合 s。
func (s *Set) Contains(i interface{}) bool {
	if len(s.data) == 0 {
		return false
	}

	s.validate(i)
	_, ok := s.data[i]
	return ok
}

// Keys 返回集合 s 中的键。
func (s *Set) Keys() []interface{} {
	var keys []interface{}

	for key := range s.data {
		keys = append(keys, key)
	}

	return keys
}

// KeysInt 返回集合 s 中的 int 键。
func (s *Set) KeysInt() []int {
	var keys []int

	for key := range s.data {
		if intKey, ok := key.(int); ok {
			keys = append(keys, intKey)
		}
	}

	return keys
}

// KeysInt64 返回集合 s 中的 int64 键。
func (s *Set) KeysInt64() []int64 {
	var keys []int64

	for key := range s.data {
		if intKey, ok := key.(int64); ok {
			keys = append(keys, intKey)
		}
	}

	return keys
}

// KeysUint 返回集合 s 中的 uint 键。
func (s *Set) KeysUint() []uint {
	var keys []uint

	for key := range s.data {
		if intKey, ok := key.(uint); ok {
			keys = append(keys, intKey)
		}
	}

	return keys
}

// KeysUint64 返回集合 s 中的 uint64 键。
func (s *Set) KeysUint64() []uint64 {
	var keys []uint64

	for key := range s.data {
		if intKey, ok := key.(uint64); ok {
			keys = append(keys, intKey)
		}
	}

	return keys
}

// KeysStr 返回集合 s 中的 string 键。
func (s *Set) KeysStr() []string {
	var keys []string

	for key := range s.data {
		if strKey, ok := key.(string); ok {
			keys = append(keys, strKey)
		}
	}

	return keys
}

// Remove 从集合 s 中移除元素 i。
func (s *Set) Remove(i interface{}) {
	s.validate(i)
	delete(s.data, i)
}

// Count 返回集合 s 中的元素个数。
func (s *Set) Count() int {
	return len(s.data)
}

func (s *Set) add(i interface{}) {
	switch s.tp {
	case unmanaged:
	// 啥也不做
	case untyped:
		s.setType(i)
	default:
		s.validate(i)
	}
	s.data[i] = lang.Placeholder
}

func (s *Set) setType(i interface{}) {
	switch i.(type) {
	case int:
		s.tp = intType
	case int64:
		s.tp = int64Type
	case uint:
		s.tp = uintType
	case uint64:
		s.tp = uintType
	case string:
		s.tp = stringType
	}
}

func (s *Set) validate(i interface{}) {
	if s.tp == unmanaged {
		return
	}

	switch i.(type) {
	case int:
		if s.tp != intType {
			logx.Errorf("错误：元素是 int，但集合可包括的元素类型为 %d", s.tp)
		}
	case int64:
		if s.tp != int64Type {
			logx.Errorf("错误：元素是 int64，但集合可包括的元素类型为 %d", s.tp)
		}
	case uint:
		if s.tp != uintType {
			logx.Errorf("错误：元素是 uint，但集合可包括的元素类型为 %d", s.tp)
		}
	case uint64:
		if s.tp != uint64Type {
			logx.Errorf("错误：元素是 uint64，但集合可包括的元素类型为 %d", s.tp)
		}
	case string:
		if s.tp != stringType {
			logx.Errorf("错误：元素是 string，但集合可包括的元素类型为 %d", s.tp)
		}

	}
}
