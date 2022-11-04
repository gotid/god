package apigen

import (
	_ "embed"
	"errors"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/console"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/spf13/cobra"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed api.tpl
var apiTemplate string

var (
	// VarStringOutput 表示输出。
	VarStringOutput string
	// VarStringHome 表示 god home 文件夹。
	VarStringHome string
	// VarStringRemote 表示 god 远程 git 仓库。
	VarStringRemote string
	// VarStringBranch 表示 god 远程 git 分支。
	VarStringBranch string
)

// CreateApiTemplate 创建 api 模板文件
func CreateApiTemplate(_ *cobra.Command, _ []string) error {
	apiFile := VarStringOutput
	if len(apiFile) == 0 {
		return errors.New("缺少 -o")
	}

	file, err := pathx.CreateIfNotExist(apiFile)
	if err != nil {
		return err
	}
	defer file.Close()

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

	baseName := pathx.FilenameWithoutExt(filepath.Base(apiFile))
	if strings.HasSuffix(strings.ToLower(baseName), "-api") {
		baseName = baseName[:len(baseName)-4]
	} else if strings.HasSuffix(strings.ToLower(baseName), "api") {
		baseName = baseName[:len(baseName)-3]
	}

	t := template.Must(template.New("etcTemplate").Parse(text))
	if err = t.Execute(file, map[string]string{
		"gitUser":     getGitUser(),
		"gitEmail":    getGitEmail(),
		"serviceName": baseName + "-api",
	}); err != nil {
		return err
	}

	console.NewColorConsole().MarkDone()

	return nil
}
