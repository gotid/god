package ctx

import (
	"errors"
	"github.com/gotid/god/tools/god/rpc/execx"
	"path/filepath"
)

var errModuleCheck = errors.New("工作目录必须存在于 go.mod 或 $GOPATH")

// ProjectContext 是一个项目结构，包括 WorkDir, Name, Path 和 Dir。
type ProjectContext struct {
	// 工作目录
	WorkDir string
	// 项目名称，如 user-center、course
	Name string
	// 项目所属模块路径。
	// 模块要么是一个 go mod 项目，要么是项目根名称。
	// 如 github.com/gotid/god 或 greet
	Path string
	// 项目所在文件目录，如 /Users/zs/goland/god
	Dir string
}

// Prepare 检查项目所属模块并返回。
// workDir 是生成代码的源文件目录。
func Prepare(workDir string) (*ProjectContext, error) {
	ctx, err := background(workDir)
	if err == nil {
		return ctx, nil
	}

	name := filepath.Base(workDir)
	_, err = execx.Run("go mod init "+name, workDir)
	if err != nil {
		return nil, err
	}

	return background(workDir)
}

func background(workDir string) (*ProjectContext, error) {
	isGoMod, err := IsGoMod(workDir)
	if err != nil {
		return nil, err
	}

	if isGoMod {
		return projectFromGoMod(workDir)
	}

	return projectFromGoPath(workDir)
}
