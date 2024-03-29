package golang

import (
	"os"
	"os/exec"
)

// Install 安装 golang 模块。
func Install(git string) error {
	cmd := exec.Command("go", "install", git)
	env := os.Environ()
	env = append(env, "GO111MODULE=on", "GOPROXY=https://goproxy.cn,direct")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}
