package ctx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gotid/god/tools/god/rpc/execx"
	"github.com/gotid/god/tools/god/util/pathx"
	"os"
	"path/filepath"
	"strings"
)

const goModuleWithoutGoFiles = "command-line-arguments"

var errInvalidGoMod = errors.New("无效的 go module")

// Module 命令 `go list` 的结果
type Module struct {
	Path      string
	Main      bool
	Dir       string
	GoMod     string
	GoVersion string
}

func (m *Module) validate() error {
	if m.Path == goModuleWithoutGoFiles || m.Dir == "" {
		return errInvalidGoMod
	}

	return nil
}

// 用于查找给定目录的 go module 和项目文件路径。
func projectFromGoMod(workDir string) (*ProjectContext, error) {
	if len(workDir) == 0 {
		return nil, errors.New("工作目录不能为空")
	}
	if _, err := os.Stat(workDir); err != nil {
		return nil, err
	}

	workDir, err := pathx.ReadLink(workDir)
	if err != nil {
		return nil, err
	}

	module, err := getRealModule(workDir, execx.Run)
	if err != nil {
		return nil, err
	}
	if err := module.validate(); err != nil {
		return nil, err
	}

	ctx := ProjectContext{
		WorkDir: workDir,
		Name:    filepath.Base(module.Dir),
		Path:    module.Path,
	}
	dir, err := pathx.ReadLink(module.Dir)
	if err != nil {
		return nil, err
	}
	ctx.Dir = dir

	return &ctx, nil
}

func getRealModule(workDir string, execRun execx.RunFunc) (*Module, error) {
	data, err := execRun("go list -json -m", workDir)
	if err != nil {
		return nil, err
	}

	modules, err := decodePackages(strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	for _, module := range modules {
		if strings.HasPrefix(workDir, module.Dir) {
			return &module, nil
		}
	}

	return nil, errors.New("没有匹配的模块")
}

func decodePackages(reader *strings.Reader) ([]Module, error) {
	var modules []Module
	decoder := json.NewDecoder(reader)
	for decoder.More() {
		var m Module
		if err := decoder.Decode(&m); err != nil {
			return nil, fmt.Errorf("无效模块：%v", err)
		}
		modules = append(modules, m)
	}

	return modules, nil
}
