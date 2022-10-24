package protocgengogrpc

import (
	"github.com/gotid/god/tools/god/pkg/god"
	"github.com/gotid/god/tools/god/pkg/golang"
	"github.com/gotid/god/tools/god/rpc/execx"
	"github.com/gotid/god/tools/god/util/env"
	"strings"
)

const (
	Name = "protoc-gen-go-grpc"
	url  = "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
)

// Install 安装 protoc-gen-go-grpc。
func Install(cacheDir string) (string, error) {
	return god.Install(cacheDir, Name, func(dest string) (string, error) {
		err := golang.Install(url)
		return dest, err
	})
}

// Exists 判断 protoc-gen-go-grpc 插件是否存在。
func Exists() bool {
	_, err := env.LookupProtocGenGoGrpc()
	return err == nil
}

// Version 返回 protoc-gen-go-grpc 插件的版本号。
func Version() (string, error) {
	path, err := env.LookupProtocGenGoGrpc()
	if err != nil {
		return "", err
	}

	version, err := execx.Run(path+" --version", "")
	if err != nil {
		return "", err
	}

	fields := strings.Fields(version)
	if len(fields) > 1 {
		return fields[1], nil
	}

	return "", nil
}
