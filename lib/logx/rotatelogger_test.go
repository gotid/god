package logx

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/gotid/god/lib/fs"

	"github.com/stretchr/testify/assert"
)

func TestDailyRotateRule_MarkRotated(t *testing.T) {
	var rule DailyRotateRule
	rule.MarkRotated()
	assert.Equal(t, getNowDate(), rule.rotatedTime)
}

func TestDailyRotateRule_OutdatedFiles(t *testing.T) {
	var rule DailyRotateRule
	assert.Empty(t, rule.OutdatedFiles())
	rule.days = 1
	assert.Empty(t, rule.OutdatedFiles())
	rule.gzip = true
	assert.Empty(t, rule.OutdatedFiles())
}

func TestDailyRotateRule_ShallRotate(t *testing.T) {
	var rule DailyRotateRule
	rule.rotatedTime = time.Now().Add(time.Hour * 24).Format(dateFormat)
	assert.True(t, rule.ShallRotate(0))
}

func TestSizeLimitRotateRule_MarkRotated(t *testing.T) {
	var rule SizeLimitRotateRule
	rule.MarkRotated()
	assert.Equal(t, getNowDateInRFC3339Format(), rule.rotatedTime)
}

func TestSizeLimitRotateRule_OutdatedFiles(t *testing.T) {
	var rule SizeLimitRotateRule
	assert.Empty(t, rule.OutdatedFiles())
	rule.days = 1
	assert.Empty(t, rule.OutdatedFiles())
	rule.gzip = true
	assert.Empty(t, rule.OutdatedFiles())
	rule.maxBackups = 0
	assert.Empty(t, rule.OutdatedFiles())
}

func TestSizeLimitRotateRule_ShallRotate(t *testing.T) {
	var rule SizeLimitRotateRule
	rule.rotatedTime = time.Now().Add(time.Hour * 24).Format(fileTimeFormat)
	rule.maxSize = 0
	assert.False(t, rule.ShallRotate(0))
	rule.maxSize = 100
	assert.False(t, rule.ShallRotate(0))
	assert.True(t, rule.ShallRotate(101*megaBytes))
}

func TestRotateLogger_Close(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filename)
	}
	logger, err := NewLogger(filename, new(DailyRotateRule), false)
	assert.Nil(t, err)
	assert.Nil(t, logger.Close())
}

func TestRotateLogger_GetBackupFilename(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filename)
	}
	logger, err := NewLogger(filename, new(DailyRotateRule), false)
	assert.Nil(t, err)
	assert.True(t, len(logger.getBackupFilename()) > 0)
	logger.backup = ""
	assert.True(t, len(logger.getBackupFilename()) > 0)
}

func TestRotateLogger_MayCompressFile(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() {
		os.Stdout = old
	}()

	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filename)
	}

	logger, err := NewLogger(filename, new(DailyRotateRule), false)
	assert.Nil(t, err)
	logger.maybeCompressFile(filename)
	_, err = os.Stat(filename)
	assert.Nil(t, err)
}

func TestRotateLogger_MayCompressFileTrue(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() {
		os.Stdout = old
	}()

	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	logger, err := NewLogger(filename, new(DailyRotateRule), true)
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filepath.Base(logger.getBackupFilename()) + ".gz")
	}
	logger.maybeCompressFile(filename)
	_, err = os.Stat(filename)
	assert.NotNil(t, err)
}

func TestRotateLogger_Rotate(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	logger, err := NewLogger(filename, new(DailyRotateRule), true)
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer func() {
			os.Remove(logger.getBackupFilename())
			os.Remove(filepath.Base(logger.getBackupFilename()) + ".gz")
		}()
	}
	err = logger.rotate()
	switch v := err.(type) {
	case *os.LinkError:
		// 避免 docker 容器上的重命名错误
		assert.Equal(t, syscall.EXDEV, v.Err)
	case *os.PathError:
		// 忽略测试的删除错误
		// 文件在 GitHub 操作中被清理。
		assert.Equal(t, "remove", v.Op)
	default:
		assert.Nil(t, err)
	}
}

func TestRotateLogger_Write(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	rule := new(DailyRotateRule)
	logger, err := NewLogger(filename, rule, true)
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer func() {
			os.Remove(logger.getBackupFilename())
			os.Remove(filepath.Base(logger.getBackupFilename()) + ".gz")
		}()
	}
	// 由于 DATA RACE，以下 write 调用无法更改为 Write。
	logger.write([]byte(`foo`))
	rule.rotatedTime = time.Now().Add(-time.Hour * 24).Format(dateFormat)
	logger.write([]byte(`bar`))
	logger.Close()
	logger.write([]byte(`baz`))
}

func TestLogWriter_Close(t *testing.T) {
	assert.Nil(t, newLogWriter(nil).Close())
}

func TestRotateLogger_WithSizeLimitRotateRuleClose(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filename)
	}
	logger, err := NewLogger(filename, new(SizeLimitRotateRule), false)
	assert.Nil(t, err)
	assert.Nil(t, logger.Close())
}

func TestRotateLogger_GetBackupWithSizeLimitRotateRuleFilename(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filename)
	}
	logger, err := NewLogger(filename, new(SizeLimitRotateRule), false)
	assert.Nil(t, err)
	assert.True(t, len(logger.getBackupFilename()) > 0)
	logger.backup = ""
	assert.True(t, len(logger.getBackupFilename()) > 0)
}

func TestRotateLogger_WithSizeLimitRotateRuleMayCompressFile(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() {
		os.Stdout = old
	}()

	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filename)
	}

	logger, err := NewLogger(filename, new(SizeLimitRotateRule), false)
	assert.Nil(t, err)
	logger.maybeCompressFile(filename)
	_, err = os.Stat(filename)
	assert.Nil(t, err)
}

func TestRotateLogger_WithSizeLimitRotateRuleMayCompressFileTrue(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() {
		os.Stdout = old
	}()

	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	logger, err := NewLogger(filename, new(SizeLimitRotateRule), true)
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer os.Remove(filepath.Base(logger.getBackupFilename()) + ".gz")
	}
	logger.maybeCompressFile(filename)
	_, err = os.Stat(filename)
	assert.NotNil(t, err)
}

func TestRotateLogger_WithSizeLimitRotateRuleRotate(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	logger, err := NewLogger(filename, new(SizeLimitRotateRule), true)
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer func() {
			os.Remove(logger.getBackupFilename())
			os.Remove(filepath.Base(logger.getBackupFilename()) + ".gz")
		}()
	}
	err = logger.rotate()
	switch v := err.(type) {
	case *os.LinkError:
		assert.Equal(t, syscall.EXDEV, v.Err)
	case *os.PathError:
		assert.Equal(t, "remove", v.Op)
	default:
		assert.Nil(t, err)
	}
}

func TestRotateLogger_WithSizeLimitRotateRuleWrite(t *testing.T) {
	filename, err := fs.TempFilenameWithText("foo")
	assert.Nil(t, err)
	rule := new(SizeLimitRotateRule)
	logger, err := NewLogger(filename, rule, true)
	assert.Nil(t, err)
	if len(filename) > 0 {
		defer func() {
			os.Remove(logger.getBackupFilename())
			os.Remove(filepath.Base(logger.getBackupFilename()) + ".gz")
		}()
	}

	logger.write([]byte(`foo`))
	rule.rotatedTime = time.Now().Add(-time.Hour * 24).Format(dateFormat)
	logger.write([]byte(`bar`))
	logger.Close()
	logger.write([]byte(`baz`))
}

func BenchmarkRotateLogger(b *testing.B) {
	filename1 := "./test1.log"
	dailyRotateRuleLogger, err1 := NewLogger(
		filename1,
		DefaultRotateRule(
			filename1,
			backupFileDelimiter,
			1,
			true,
		),
		true,
	)
	if err1 != nil {
		b.Logf("未能新建每日轮换规则日志记录器：%v", err1)
		b.FailNow()
	}

	filename2 := "./test2.log"
	sizeLimitRotateRuleLogger, err2 := NewLogger(
		filename2,
		NewSizeLimitRotateRule(
			filename2,
			backupFileDelimiter,
			1,
			100,
			10,
			true,
		),
		true,
	)
	if err2 != nil {
		b.Logf("未能新建大小受限的轮换规则日志记录器：%v", err2)
		b.FailNow()
	}

	defer func() {
		dailyRotateRuleLogger.Close()
		sizeLimitRotateRuleLogger.Close()
		os.Remove(filename1)
		os.Remove(filename2)
	}()

	b.Run("每日轮换规则", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dailyRotateRuleLogger.write([]byte("测试中\n测试中\n"))
		}
	})

	b.Run("大小受限的轮换规则", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sizeLimitRotateRuleLogger.write([]byte("测试中\n测试中\n"))
		}
	})
}
