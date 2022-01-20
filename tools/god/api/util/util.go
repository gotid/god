package util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/gotid/god/lib/fs"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/tools/god/api/spec"
)

func MaybeCreateFile(dir, subdir, file string) (fp *os.File, created bool, err error) {
	logx.Must(fs.MkdirIfNotExist(path.Join(dir, subdir)))
	fpath := path.Join(dir, subdir, file)
	if fs.FileExist(fpath) {
		fmt.Printf("%s 已存在，忽略生成\n", fpath)
		return nil, false, nil
	}

	fp, err = fs.CreateIfNotExist(fpath)
	created = err == nil
	return
}

func WrapErr(err error, message string) error {
	return errors.New(message + ", " + err.Error())
}

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
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

func ComponentName(api *spec.ApiSpec) string {
	name := api.Service.Name
	if strings.HasSuffix(name, "-api") {
		return name[:len(name)-4] + "Components"
	}
	return name + "Components"
}

func WriteIndent(writer io.Writer, indent int) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(writer, "\t")
	}
}

func RemoveComment(line string) string {
	commentIdx := strings.Index(line, "//")
	if commentIdx >= 0 {
		return strings.TrimSpace(line[:commentIdx])
	}
	return strings.TrimSpace(line)
}
