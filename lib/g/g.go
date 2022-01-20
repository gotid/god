package g

import "github.com/gotid/god/lib/container/gvar"

// Var 是通用变量接口，与泛型类似。
type Var = gvar.Var

// Map 类型别名。
type (
	Map        = map[string]interface{}
	MapAnyAny  = map[interface{}]interface{}
	MapAnyStr  = map[interface{}]string
	MapAnyInt  = map[interface{}]int
	MapStrAny  = map[string]interface{}
	MapStrStr  = map[string]string
	MapStrInt  = map[string]int
	MapIntAny  = map[int]interface{}
	MapIntStr  = map[int]string
	MapIntInt  = map[int]int
	MapAnyBool = map[interface{}]bool
	MapStrBool = map[string]bool
	MapIntBool = map[int]bool
)

// 常用的 List 类型别名。
type (
	List        = []Map
	ListAnyAny  = []Map
	ListAnyStr  = []MapAnyStr
	ListAnyInt  = []MapAnyInt
	ListStrAny  = []MapStrAny
	ListStrStr  = []MapStrStr
	ListStrInt  = []MapStrInt
	ListIntAny  = []MapIntAny
	ListIntStr  = []MapIntStr
	ListIntInt  = []MapIntInt
	ListAnyBool = []MapAnyBool
	ListStrBool = []MapStrBool
	ListIntBool = []MapIntBool
)

// 常用的 Slice 类型别名。
type (
	Slice    = []interface{}
	SliceAny = []interface{}
	SliceStr = []string
	SliceInt = []int
)

// Array 是 Slice 别名。
type (
	Array    = []interface{}
	ArrayAny = []interface{}
	ArrayStr = []string
	ArrayInt = []int
)
