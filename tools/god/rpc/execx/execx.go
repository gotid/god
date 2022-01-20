package execx

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gotid/god/lib/fs"

	"github.com/gotid/god/tools/god/vars"
)

func Run(arg string, dir string) (string, error) {
	goos := runtime.GOOS
	var cmd *exec.Cmd
	switch goos {
	case vars.OsMac, vars.OsLinux:
		cmd = exec.Command("sh", "-c", arg)
	case vars.OsWindows:
		cmd = exec.Command("cmd.exe", "/c", arg)
	default:
		return "", fmt.Errorf("unexpected os: %v", goos)
	}
	if len(dir) > 0 {
		cmd.Dir = dir
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(strings.TrimSuffix(stderr.String(), fs.NL))
		}
		return "", err
	}

	return strings.TrimSuffix(stdout.String(), fs.NL), nil
}
