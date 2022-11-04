package pathx

import (
	"os"
	"strings"
)

const (
	pkgSep           = "/"
	goModeIdentifier = "go.mod"
)

// JoinPackages 调用 strings.Join 并返回
func JoinPackages(pkgs ...string) string {
	return strings.Join(pkgs, pkgSep)
}

// MkdirIfNotExist 若目录不存在则创建。
func MkdirIfNotExist(dir string) error {
	if len(dir) == 0 {
		return nil
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}

	return nil

}

// 判断给定的文件名是否为软连接。
func isLink(name string) (bool, error) {
	fi, err := os.Lstat(name)
	if err != nil {
		return false, err
	}

	return fi.Mode()&os.ModeSymlink != 0, nil
}
