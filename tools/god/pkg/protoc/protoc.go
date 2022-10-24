package protoc

import (
	"archive/zip"
	"fmt"
	"github.com/gotid/god/tools/god/pkg/downloader"
	"github.com/gotid/god/tools/god/pkg/god"
	"github.com/gotid/god/tools/god/rpc/execx"
	"github.com/gotid/god/tools/god/util/env"
	"github.com/gotid/god/tools/god/util/zipx"
	"github.com/gotid/god/tools/god/vars"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	Name        = "protoc"
	ZipFilename = Name + ".zip"
)

var url = map[string]string{
	"linux_32":   "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-linux-x86_32.zip",
	"linux_64":   "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-linux-x86_64.zip",
	"darwin":     "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-osx-x86_64.zip",
	"windows_32": "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-win32.zip",
	"windows_64": "https://github.com/protocolbuffers/protobuf/releases/download/v3.19.4/protoc-3.19.4-win64.zip",
}

// Install 安装 protoc。
func Install(cacheDir string) (string, error) {
	return god.Install(cacheDir, Name, func(dest string) (string, error) {
		goos := runtime.GOOS
		tempFile := filepath.Join(os.TempDir(), ZipFilename)
		bit := 32 << (^uint(0) >> 63)
		var downloadUrl string
		switch goos {
		case vars.OsMac:
			downloadUrl = url[vars.OsMac]
		case vars.OsWindows:
			downloadUrl = url[fmt.Sprintf("%s_%d", vars.OsWindows, bit)]
		case vars.OsLinux:
			downloadUrl = url[fmt.Sprintf("%s_%d", vars.OsLinux, bit)]
		default:
			return "", fmt.Errorf("不支持的操作系统：%q", goos)
		}

		err := downloader.Download(downloadUrl, tempFile)
		if err != nil {
			return "", err
		}

		return dest, zipx.Unpacking(tempFile, filepath.Dir(dest), func(f *zip.File) bool {
			return filepath.Base(f.Name) == filepath.Base(dest)
		})
	})
}

// Exists 判断 protoc 是否存在。
func Exists() bool {
	_, err := env.LookupProtoc()
	return err == nil
}

// Version 返回 protoc 版本号。
func Version() (string, error) {
	path, err := env.LookupProtoc()
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
