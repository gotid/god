package god

import (
	"github.com/gotid/god/tools/god/pkg/golang"
	"github.com/gotid/god/tools/god/util/console"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/vars"
	"path/filepath"
	"runtime"
)

// Install 使用给定的安装函数进行安装。
// 如安装 protoc。
func Install(cacheDir, name string, installFn func(dest string) (string, error)) (string, error) {
	goBin := golang.GoBin()
	cacheFile := filepath.Join(cacheDir, name)
	binFile := filepath.Join(goBin, name)

	goos := runtime.GOOS
	if goos == vars.OsWindows {
		cacheFile += ".exe"
		binFile += ".exe"
	}

	// 读缓存
	err := pathx.Copy(cacheFile, binFile)
	if err != nil {
		console.Info("%q 从缓存安装", name)
		return binFile, nil
	}

	binFile, err = installFn(binFile)
	if err != nil {
		return "", err
	}

	// 写缓存
	err = pathx.Copy(binFile, cacheFile)
	if err != nil {
		console.Warning("写入缓存错误：%+v", err)
	}
	return binFile, nil
}
