package util

import (
	"fmt"
	"github.com/gotid/god/tools/god/util/env"
	"github.com/gotid/god/tools/god/util/pathx"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// CloneIntoGitHome 克隆远程分支。
func CloneIntoGitHome(url, branch string) (dir string, err error) {
	gitHome, err := pathx.GetGitHome()
	if err != nil {
		return "", err
	}

	os.RemoveAll(gitHome)
	ext := filepath.Ext(url)
	repo := strings.TrimSuffix(filepath.Base(url), ext)
	dir = filepath.Join(gitHome, repo)
	if pathx.FileExists(dir) {
		os.RemoveAll(dir)
	}

	if !env.CanExec() {
		return "", fmt.Errorf("系统 %q 无法调用 'exec' 命令", runtime.GOOS)
	}

	path, err := env.LookPath("git")
	if err != nil {
		return "", err
	}
	args := []string{"clone"}
	if len(branch) > 0 {
		args = append(args, "-b", branch)
	}
	args = append(args, url, dir)
	cmd := exec.Command(path, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return
}
