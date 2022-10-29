package fs

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestTempFileWithText(t *testing.T) {
	f, err := TempFileWithText("hi")
	if err != nil {
		t.Error(err)
	}
	if f == nil {
		t.Error("TempFileWithText 返回了空文件")
	}
	if f.Name() == "" {
		t.Error("TempFileWithText 返回了空白的文件名称")
	}
	defer os.Remove(f.Name())

	bs, err := io.ReadAll(f)
	assert.Nil(t, err)
	if len(bs) != 4 {
		t.Error("TempFileWithText 返回了错误的文件大小")
	}
	if f.Close() != nil {
		t.Error("TempFileWithText 在关闭时发生错误")
	}
}

func TestTempFilenameWithText(t *testing.T) {
	f, err := TempFilenameWithText("hi")
	if err != nil {
		t.Error(err)
	}
	if f == "" {
		t.Error("TempFilenameWithText 返回了空白的文件名称")
	}
	defer os.Remove(f)

	bs, err := os.ReadFile(f)
	assert.Nil(t, err)
	if len(bs) != 4 {
		t.Error("TempFilenameWithText 返回了错误的文件大小")
	}
}
