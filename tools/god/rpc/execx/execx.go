package execx

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/vars"
	"os/exec"
	"runtime"
	"strings"
)

// RunFunc 定义一个 Run 函数的函数类型。
type RunFunc func(string, string, ...*bytes.Buffer) (string, error)

// Run 返回 golang 中执行 shell 脚本的结果。
// 支持 macOS，windows和linux 操作系统，其他系统暂不支持。
func Run(arg, dir string, in ...*bytes.Buffer) (string, error) {
	goos := runtime.GOOS
	var cmd *exec.Cmd
	switch goos {
	case vars.OsMac, vars.OsLinux:
		cmd = exec.Command("sh", "-c", arg)
	case vars.OsWindows:
		cmd = exec.Command("cmd.exe", "/c", arg)
	default:
		return "", fmt.Errorf("暂不支持的操作系统：%v", goos)
	}

	if len(dir) > 0 {
		cmd.Dir = dir
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if len(in) > 0 {
		cmd.Stdin = in[0]
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(strings.TrimSuffix(stderr.String(), pathx.NL))
		}
		return "", err
	}

	return strings.TrimSuffix(stdout.String(), pathx.NL), nil
}
