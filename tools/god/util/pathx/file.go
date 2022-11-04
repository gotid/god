package pathx

import (
	"fmt"
	"github.com/gotid/god/tools/god/internal/version"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	NL       = "\n"
	godDir   = ".god"
	gitDir   = ".git"
	cacheDir = "cache"
)

var godHome string

// RegisterGodHome 注册 god 主目录。
func RegisterGodHome(home string) {
	godHome = home
}

// LoadTemplate 获取指定模板文件的内容。
// 如果模板文件不存在，则返回内置模板文本。
func LoadTemplate(category, file, builtin string) (string, error) {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return "", err
	}

	file = filepath.Join(dir, file)
	if !FileExists(file) {
		return builtin, err
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// GetGodHome 返回 god 代码生成器的路径，默认路径是 ~/.god。
// 可通过调用 RegisterGodHome 方法自定义该路径。
func GetGodHome() (home string, err error) {
	defer func() {
		if err != nil {
			return
		}
		info, err := os.Stat(home)
		if err == nil && !info.IsDir() {
			os.Rename(home, home+".old")
			MkdirIfNotExist(home)
		}
	}()

	if len(godHome) != 0 {
		home = godHome
		return
	}

	home, err = GetDefaultGodHome()
	return
}

// GetDefaultGodHome 返回 god 代码生成器的默认用户主目录路径。
// 默认路径为 $HOME/.god
func GetDefaultGodHome() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, godDir), nil
}

// GetGitHome 获取 god 代码生成器的 git 主目录。
func GetGitHome() (string, error) {
	homeDir, err := GetGodHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, gitDir), nil
}

// GetCacheDir 获取 god 代码生成器的缓存目录。
func GetCacheDir() (string, error) {
	homeDir, err := GetGodHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, cacheDir), nil
}

// FileExists 判断给定的文件是否存在。
func FileExists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

// FilenameWithoutExt 返回一个没有扩展名的文件名。
func FilenameWithoutExt(file string) string {
	return strings.TrimSuffix(file, filepath.Ext(file))
}

// SameFile 比较两个路径是否为相同路径。 如：/Users/god 与 /Users/God
func SameFile(path1, path2 string) (bool, error) {
	stat1, err := os.Stat(path1)
	if err != nil {
		return false, err
	}

	stat2, err := os.Stat(path2)
	if err != nil {
		return false, err
	}

	return os.SameFile(stat1, stat2), nil
}

// CreateIfNotExist 文件如不存在则创建。
func CreateIfNotExist(file string) (*os.File, error) {
	_, err := os.Stat(file)
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("文件 %s 已经存在", file)
	}

	return os.Create(file)
}

// MustTempDir 创建一个临时文件夹。
func MustTempDir() string {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		log.Fatalln(err)
	}

	return dir
}

// GetTemplateDir 通过 GetGodHome 获取给定的类别路径。
func GetTemplateDir(category string) (string, error) {
	home, err := GetGodHome()
	if err != nil {
		return "", err
	}

	if home == godHome {
		return filepath.Join(home, category), nil
	}

	return filepath.Join(home, version.GetGodVersion(), category), nil
}

func Copy(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	dir := filepath.Dir(dest)
	err = MkdirIfNotExist(dir)
	if err != nil {
		return err
	}

	w, err := os.Create(dest)
	if err != nil {
		return err
	}
	w.Chmod(os.ModePerm)
	defer w.Close()

	_, err = io.Copy(w, f)
	return err
}
