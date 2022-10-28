package mapping

type (
	// Valuer 接口定义了使用给定键从底层对象获取值的方法。
	Valuer interface {
		// Value 获取给定键关联的值。
		Value(key string) (interface{}, bool)
	}

	// MapValuer 是一个使用 Value 方法获取给定键的值的字典。
	MapValuer map[string]interface{}
)

// Value 从字典 mv 中获取给定键的值。
func (mv MapValuer) Value(key string) (interface{}, bool) {
	v, ok := mv[key]
	return v, ok
}
