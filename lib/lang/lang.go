package lang

type (
	// AnyType 用于保存任何类型。
	AnyType = interface{}
	// PlaceholderType 表示一个占位符类型。
	PlaceholderType = struct{}
)

// Placeholder 是一个全局可用的占位符对象。
var Placeholder PlaceholderType
