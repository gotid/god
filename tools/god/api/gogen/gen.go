package gogen

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gotid/god/lib/logx"
	apiFormat "github.com/gotid/god/tools/god/api/format"
	"github.com/gotid/god/tools/god/api/parser"
	apiUtil "github.com/gotid/god/tools/god/api/util"

	"github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/pkg/golang"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

const tmpFile = "%s-%d"

var (
	tmpDir = path.Join(os.TempDir(), "god")

	// VarStringDir 目录。
	VarStringDir string
	// VarStringAPI API。
	VarStringAPI string
	// VarStringHome god 主目录。
	VarStringHome string
	// VarStringRemote 表示 god 远程 git 仓库。
	VarStringRemote string
	// VarStringBranch 表示 god 远程 git 分支。
	VarStringBranch string
	// VarStringStyle 表示输出文件的命名风格。
	VarStringStyle string
)

// GoCommand 根据 api 协议文件，生成 api 示例服务
func GoCommand(_ *cobra.Command, _ []string) error {
	apiFile := VarStringAPI
	dir := VarStringDir
	namingStyle := VarStringStyle
	home := VarStringHome
	remote := VarStringRemote
	branch := VarStringBranch
	if len(remote) > 0 {
		repo, _ := util.CloneIntoGitHome(remote, branch)
		if len(repo) > 0 {
			home = repo
		}
	}

	if len(home) > 0 {
		pathx.RegisterGodHome(home)
	}
	if len(apiFile) == 0 {
		return errors.New("缺失 -api")
	}
	if len(dir) == 0 {
		return errors.New("缺失 -dir")
	}

	return DoGenProject(apiFile, dir, namingStyle)
}

// DoGenProject 通过给定的 api 协议文件生成 go 项目。
func DoGenProject(apiFile, dir, style string) error {
	api, err := parser.Parse(apiFile)
	if err != nil {
		return err
	}

	if err = api.Validate(); err != nil {
		return err
	}

	cfg, err := config.NewConfig(style)
	if err != nil {
		return err
	}

	logx.Must(pathx.MkdirIfNotExist(dir))
	rootPkg, err := golang.GetParentPackage(dir)
	if err != nil {
		return err
	}

	logx.Must(genEtc(dir, cfg, api))
	logx.Must(genConfig(dir, cfg, api))
	logx.Must(genMain(dir, rootPkg, cfg, api))
	logx.Must(genServiceContext(dir, rootPkg, cfg, api))
	logx.Must(genTypes(dir, cfg, api))
	logx.Must(genRoutes(dir, rootPkg, cfg, api))
	logx.Must(genHandlers(dir, rootPkg, cfg, api))
	logx.Must(genLogic(dir, rootPkg, cfg, api))
	logx.Must(genMiddleware(dir, cfg, api))

	if err = backupAndSweep(apiFile); err != nil {
		return err
	}

	if err = apiFormat.ApiFormatByPath(apiFile, false); err != nil {
		return err
	}

	fmt.Println(aurora.Green("完成。"))
	return nil
}

func backupAndSweep(apiFile string) error {
	var err error
	var wg sync.WaitGroup

	wg.Add(2)
	_ = os.MkdirAll(tmpDir, os.ModePerm)

	go func() {
		_, fileName := filepath.Split(apiFile)
		_, e := apiUtil.Copy(apiFile, fmt.Sprintf(path.Join(tmpDir, tmpFile), fileName, time.Now().Unix()))
		if e != nil {
			err = e
		}
		wg.Done()
	}()
	go func() {
		if e := sweep(); e != nil {
			err = e
		}
		wg.Done()
	}()
	wg.Wait()

	return err
}

func sweep() error {
	keepTime := time.Now().AddDate(0, 0, -7)
	return filepath.Walk(tmpDir, func(fpath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		pos := strings.LastIndexByte(info.Name(), '-')
		if pos > 0 {
			timestamp := info.Name()[pos+1:]
			seconds, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				// print error and ignore
				fmt.Println(aurora.Red(fmt.Sprintf("扫描忽略的文件：%s", fpath)))
				return nil
			}

			tm := time.Unix(seconds, 0)
			if tm.Before(keepTime) {
				if err := os.Remove(fpath); err != nil {
					fmt.Println(aurora.Red(fmt.Sprintf("无法删除文件：%s", fpath)))
					return err
				}
			}
		}

		return nil
	})
}
