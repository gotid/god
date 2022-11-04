package util

import (
	"errors"
	"fmt"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/tools/god/api/spec"
	"github.com/gotid/god/tools/god/util/pathx"
	"io"
	"os"
	"path"
	"strings"
)

// MaybeCreateFile 若文件不存在则创建。
func MaybeCreateFile(dir, subDir, file string) (fp *os.File, created bool, err error) {
	logx.Must(pathx.MkdirIfNotExist(path.Join(dir, subDir)))
	filePath := path.Join(dir, subDir, file)
	if pathx.FileExists(filePath) {
		fmt.Printf("%s 已存在，不再生成\n", filePath)
		return nil, false, nil
	}

	fp, err = pathx.CreateIfNotExist(filePath)
	created = err == nil
	return
}

// WrapErr 用给定消息包装一个错误。
func WrapErr(err error, message string) error {
	return errors.New(message + ", " + err.Error())
}

// Copy 如果源文件和目标文件存在，则调用 io.Copy。
func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s 不是一个常规文件", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// ComponentName 返回 typescript 的组件名称
func ComponentName(api *spec.ApiSpec) string {
	name := api.Service.Name
	if strings.HasSuffix(name, "-api") {
		return name[:len(name)-4] + "Components"
	}
	return name + "Components"
}

// WriteIndent 写入制表符缩进空间。
func WriteIndent(writer io.Writer, indent int) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(writer, "\t")
	}
}

// RemoveComment 移除评论内容。
func RemoveComment(line string) string {
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		return strings.TrimSpace(line[:commentIdx])
	}
	return strings.TrimSpace(line)
}
