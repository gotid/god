package lang

type (
	// AnyType 用于保存任意类型。
	AnyType = interface{}
	// PlaceholderType 表示一个占位符类型。
	PlaceholderType = struct{}
)

// Placeholder 是一个可全局使用的占位符对象。
var Placeholder PlaceholderType
