//go:build linux || darwin

package pathx

import (
	"os"
	"path/filepath"
)

// ReadLink 返回命名符号链接的递归目标。
func ReadLink(name string) (string, error) {
	name, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}

	if _, err := os.Lstat(name); err != nil {
		return name, err
	}

	if name == "/" || name == "/var" {
		return name, err
	}

	link, err := isLink(name)
	if err != nil {
		return "", err
	}

	if !link {
		dir, base := filepath.Split(name)
		dir = filepath.Clean(dir)
		dir, err := ReadLink(dir)
		if err != nil {
			return "", err
		}

		return filepath.Join(dir, base), nil
	}

	linked, err := os.Readlink(name)
	if err != nil {
		return "", err
	}

	dir, base := filepath.Split(linked)
	dir = filepath.Dir(dir)
	dir, err = ReadLink(dir)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, base), nil
}
