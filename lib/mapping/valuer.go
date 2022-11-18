package mapping

type (
	// Valuer 接口定义了使用给定键从底层对象获取值的方法。
	Valuer interface {
		// Value 获取给定键关联的值。
		Value(key string) (any, bool)
	}

	// MapValuer 是一个使用 Value 方法获取给定键的值的字典。
	mapValuer map[string]any

	// A valuerWithParent defines a node that has a parent node.
	valuerWithParent interface {
		Valuer
		// Parent get the parent valuer for current node.
		Parent() valuerWithParent
	}

	// A node is a map that can use Value method to get values with given keys.
	node struct {
		current Valuer
		parent  valuerWithParent
	}

	// A valueWithParent is used to wrap the value with its parent.
	valueWithParent struct {
		value  any
		parent valuerWithParent
	}

	// simpleValuer is a type to get value from current node.
	simpleValuer node

	// recursiveValuer is a type to get the value recursively from current and parent nodes.
	recursiveValuer node
)

// Value 从字典 mv 中获取给定键的值。
func (mv mapValuer) Value(key string) (any, bool) {
	v, ok := mv[key]
	return v, ok
}

// Value gets the value associated with the given key from sv.
func (sv simpleValuer) Value(key string) (any, bool) {
	v, ok := sv.current.Value(key)
	return v, ok
}

// Parent get the parent valuer from sv.
func (sv simpleValuer) Parent() valuerWithParent {
	if sv.parent == nil {
		return nil
	}

	return recursiveValuer{
		current: sv.parent,
		parent:  sv.parent.Parent(),
	}
}

// Value gets the value associated with the given key from rv,
// and it will inherit the value from parent nodes.
func (rv recursiveValuer) Value(key string) (any, bool) {
	val, ok := rv.current.Value(key)
	if !ok {
		if parent := rv.Parent(); parent != nil {
			return parent.Value(key)
		}

		return nil, false
	}

	if vm, ok := val.(map[string]any); ok {
		if parent := rv.Parent(); parent != nil {
			pv, pok := parent.Value(key)
			if pok {
				if pm, ok := pv.(map[string]any); ok {
					for k, v := range vm {
						pm[k] = v
					}
					return pm, true
				}
			}
		}
	}

	return val, true
}

// Parent get the parent valuer from rv.
func (rv recursiveValuer) Parent() valuerWithParent {
	if rv.parent == nil {
		return nil
	}

	return recursiveValuer{
		current: rv.parent,
		parent:  rv.parent.Parent(),
	}
}
