package spec

// Name 返回一个基础字符串，如 int32，int64
func (t PrimitiveType) Name() string {
	return t.RawName
}

// Comments 返回结构体的注释
func (t PrimitiveType) Comments() []string {
	return nil
}

// Documents 返回结构体的文档
func (t PrimitiveType) Documents() []string {
	return nil
}

// Name 返回结构体名称，如 User
func (t DefineStruct) Name() string {
	return t.RawName
}

// Comments 返回结构体注释
func (t DefineStruct) Comments() []string {
	return nil
}

// Documents 返回结构体文档
func (t DefineStruct) Documents() []string {
	return t.Docs
}

// Name 返回一个字典字符串，如 map[string]int
func (t MapType) Name() string {
	return t.RawName
}

// Comments returns the comments of struct
func (t MapType) Comments() []string {
	return nil
}

// Documents returns the documents of struct
func (t MapType) Documents() []string {
	return nil
}

// Name returns a slice string, such as []int
func (t ArrayType) Name() string {
	return t.RawName
}

// Comments returns the comments of struct
func (t ArrayType) Comments() []string {
	return nil
}

// Documents returns the documents of struct
func (t ArrayType) Documents() []string {
	return nil
}

// Name returns a pointer string, such as *User
func (t PointerType) Name() string {
	return t.RawName
}

// Comments returns the comments of struct
func (t PointerType) Comments() []string {
	return nil
}

// Documents returns the documents of struct
func (t PointerType) Documents() []string {
	return nil
}

// Name returns a interface string, Its fixed value is interface{}
func (t InterfaceType) Name() string {
	return t.RawName
}

// Comments returns the comments of struct
func (t InterfaceType) Comments() []string {
	return nil
}

// Documents returns the documents of struct
func (t InterfaceType) Documents() []string {
	return nil
}
