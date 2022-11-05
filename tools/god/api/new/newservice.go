package new

import (
	_ "embed"
	"errors"
	"github.com/gotid/god/tools/god/api/gogen"
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed api.tpl
var apiTemplate string

var (
	// VarStringHome god 主目录。
	VarStringHome string
	// VarStringRemote 表示 god 远程 git 仓库。
	VarStringRemote string
	// VarStringBranch 表示 god 远程 git 分支。
	VarStringBranch string
	// VarStringStyle 表示输出文件的命名风格。
	VarStringStyle string
)

// CreateServiceCommand 快速创建服务
func CreateServiceCommand(args []string) error {
	dirName := args[0]
	if len(VarStringStyle) == 0 {
		VarStringStyle = conf.DefaultFormat
	}
	if strings.Contains(dirName, "-") {
		return errors.New("api new 命令中服务名称不支持删除线，因为这将由函数名使用")
	}

	abs, err := filepath.Abs(dirName)
	if err != nil {
		return err
	}

	err = pathx.MkdirIfNotExist(abs)
	if err != nil {
		return err
	}

	dirName = filepath.Base(filepath.Clean(abs))
	filename := dirName + ".api"
	apiFilePath := filepath.Join(abs, filename)
	fp, err := os.Create(apiFilePath)
	if err != nil {
		return err
	}

	defer fp.Close()

	if len(VarStringRemote) > 0 {
		repo, _ := util.CloneIntoGitHome(VarStringRemote, VarStringBranch)
		if len(repo) > 0 {
			VarStringHome = repo
		}
	}

	if len(VarStringHome) > 0 {
		pathx.RegisterGodHome(VarStringHome)
	}

	text, err := pathx.LoadTemplate(category, apiTemplateFile, apiTemplate)
	if err != nil {
		return err
	}

	t := template.Must(template.New("template").Parse(text))
	if err := t.Execute(fp, map[string]string{
		"name":    dirName,
		"handler": strings.Title(dirName),
	}); err != nil {
		return err
	}

	err = gogen.DoGenProject(apiFilePath, abs, VarStringStyle)
	return err
}
