package fs

import (
	"github.com/gotid/god/lib/hash"
	"os"
)

// TempFileWithText 创建指定 text 的临时文件并返回已打开的文件实例。
// 因为文件为打开状态，所以调用方需要负责关闭文件并删除。
func TempFileWithText(text string) (*os.File, error) {
	file, err := os.CreateTemp(os.TempDir(), hash.Md5Hex([]byte(text)))
	if err != nil {
		return nil, err
	}

	if err = os.WriteFile(file.Name(), []byte(text), os.ModeTemporary); err != nil {
		return nil, err
	}

	return file, nil
}

// TempFilenameWithText 创建指定 text 的临时文件并返回全路径文件名。
// 调用方需要在使用后删除该文件。
func TempFilenameWithText(text string) (string, error) {
	file, err := TempFileWithText(text)
	if err != nil {
		return "", err
	}

	filename := file.Name()
	if err = file.Close(); err != nil {
		return "", err
	}

	return filename, nil
}
