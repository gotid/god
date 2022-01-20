package fs

import (
	"io/ioutil"
	"os"

	"github.com/gotid/god/lib/hash"
)

// TempFileWithText 创建带有给定内容的临时文件，
// 并返回打开的 *os.File 实例。
// 调用者应该关闭文件句柄，并按名称删除文件。
func TempFileWithText(text string) (*os.File, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), hash.Md5Hex([]byte(text)))
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(tempFile.Name(), []byte(text), os.ModeTemporary); err != nil {
		return nil, err
	}

	return tempFile, nil
}

// TempFilenameWithText 用给定的内容创建文件，并返回文件名(完整路径)。
// 调用者应随后删除。
func TempFilenameWithText(text string) (string, error) {
	tempFile, err := TempFileWithText(text)
	if err != nil {
		return "", err
	}

	filename := tempFile.Name()
	if err = tempFile.Close(); err != nil {
		return "", err
	}

	return filename, nil
}
