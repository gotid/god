package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gotid/god/lib/fs"
)

const goctlDir = ".goctl"

func GetGoctlHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, goctlDir), nil
}

func GetTemplateDir(category string) (string, error) {
	goctlHome, err := GetGoctlHome()
	if err != nil {
		return "", err
	}

	return filepath.Join(goctlHome, category), nil
}

func InitTemplates(category string, templates map[string]string) error {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return err
	}

	if err := fs.MkdirIfNotExist(dir); err != nil {
		return err
	}

	for k, v := range templates {
		if err := createTemplate(filepath.Join(dir, k), v, false); err != nil {
			return err
		}
	}

	return nil
}

func CreateTemplate(category, name, content string) error {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return err
	}
	return createTemplate(filepath.Join(dir, name), content, true)
}

func Clean(category string) error {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func LoadTemplate(category, file, builtin string) (string, error) {
	dir, err := GetTemplateDir(category)
	if err != nil {
		return "", err
	}

	file = filepath.Join(dir, file)
	if !fs.FileExist(file) {
		return builtin, nil
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func createTemplate(file, content string, force bool) error {
	if fs.FileExist(file) && !force {
		return nil
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}
