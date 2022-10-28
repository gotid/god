package util

const (
	// DontEditHead 添加文件开头声明，提示用户不要修改。
	DontEditHead = "// 代码由 god 生成，不要修改。"

	headTemplate = DontEditHead + `
// 源文件: {{.source}}
`
)

// GetHead 返回带有源文件名称的文件头。
func GetHead(source string) string {
	buffer, _ := With("head").Parse(headTemplate).Execute(map[string]interface{}{
		"source": source,
	})

	return buffer.String()
}
