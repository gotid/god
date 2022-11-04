package spec

// RoutePrefixKey 是路由的前缀关键字
const RoutePrefixKey = "pref"

type (
	// ApiSpec api 规范
	ApiSpec struct {
		Info    Info
		Syntax  ApiSyntax
		Imports []Import
		Types   []Type
		Service Service
	}

	// Info 信息语法块
	Info struct {
		Properties map[string]string
	}

	// ApiSyntax api 语法
	ApiSyntax struct {
		Version string
		Doc     Doc
		Comment Doc
	}

	// Doc 描述文档
	Doc []string

	// Import 导包语法块
	Import struct {
		Value   string
		Doc     Doc
		Comment Doc
	}

	// Type 接口定义 api 类型
	Type interface {
		Name() string
		Comments() []string
		Documents() []string
	}

	// Service 描述一个 api 服务
	Service struct {
		Name   string
		Groups []Group
	}

	// Group 描述一组路由信息
	Group struct {
		Annotation Annotation
		Routes     []Route
	}

	// Annotation kv 声明属性
	Annotation struct {
		Properties map[string]string
	}

	Route struct {
		AtServerAnnotation Annotation
		Method             string
		Path               string
		RequestType        Type
		ResponseType       Type
		Docs               Doc
		Handler            string
		AtDoc              AtDoc
		HandlerDoc         Doc
		HandlerComment     Doc
		Doc                Doc
		Comment            Doc
	}

	// AtDoc 描述一个 api 语法的元数据：@doc(...)
	AtDoc struct {
		Properties map[string]string
		Text       string
	}

	// ArrayType 描述一个 api 切片
	ArrayType struct {
		RawName string
		Value   Type
	}

	// DefineStruct 描述 api 结构
	DefineStruct struct {
		RawName string
		Members []Member
		Docs    Doc
	}

	// Member 描述一个结构的字段
	Member struct {
		Name     string
		Type     Type
		Tag      string
		Comment  string
		Docs     Doc
		IsInline bool
	}

	// InterfaceType 描述一个 api 的接口
	InterfaceType struct {
		RawName string
	}

	// MapType 描述一个 api 的字典
	MapType struct {
		RawName string // 只支持 PrimitiveType
		Key     string
		// 可以断言为 PrimitiveType: int、bool、
		// 指针类型: *string、*User、
		// 字典类型: map[${PrimitiveType}]interface、
		// 数组类型:[]int、[]User、[]*User
		// 接口类型: interface{}
		// Type
		Value Type
	}

	// PrimitiveType 描述了基本的 golang 类型，如 bool, int32, ...
	PrimitiveType struct {
		RawName string
	}

	// PointerType 描述了 api 指针类型
	PointerType struct {
		RawName string
		Type    Type
	}
)
