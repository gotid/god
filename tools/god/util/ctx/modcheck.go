package ctx

import (
	"fmt"
	"github.com/gotid/god/tools/god/rpc/execx"
	"os"
)

// IsGoMod 用于检查 workDir 是否为一个 go module 项目。
// 检查命令：`go list -json -m`
func IsGoMod(workDir string) (bool, error) {
	if len(workDir) == 0 {
		return false, fmt.Errorf("工作目录未找到 - %s", workDir)
	}

	if _, err := os.Stat(workDir); err != nil {
		return false, err
	}

	data, err := execx.Run("go list -m -f '{{.GoMod}}'", workDir)
	if err != nil || len(data) == 0 {
		return false, nil
	}

	return true, nil
}
