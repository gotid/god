package protocgengo

import (
	"github.com/gotid/god/tools/god/pkg/god"
	"github.com/gotid/god/tools/god/pkg/golang"
	"github.com/gotid/god/tools/god/rpc/execx"
	"github.com/gotid/god/tools/god/util/env"
	"strings"
	"time"
)

const (
	Name = "protoc-gen-go"
	url  = "google.golang.org/protobuf/cmd/protoc-gen-go@latest"
)

// Install 安装 protoc-gen-go。
func Install(cacheDir string) (string, error) {
	return god.Install(cacheDir, Name, func(dest string) (string, error) {
		err := golang.Install(url)
		return dest, err
	})
}

// Exists 判断 protoc-gen-go 是否存在。
func Exists() bool {
	ver, err := Version()
	if err != nil {
		return false
	}

	return len(ver) > 0
}

// Version 返回 protoc-gen-go 版本号。
// 由于老版本不支持获取版本信息导致进程阻塞，故使用计时器以防阻塞。
func Version() (string, error) {
	path, err := env.LookupProtocGenGo()
	if err != nil {
		return "", err
	}

	versionC := make(chan string)
	go func(c chan string) {
		version, _ := execx.Run(path+" --version", "")
		fields := strings.Fields(version)
		if len(fields) > 1 {
			c <- fields[1]
		}
	}(versionC)
	t := time.NewTimer(time.Second)
	select {
	case <-t.C:
		return "", nil
	case version := <-versionC:
		return version, nil
	}
}
