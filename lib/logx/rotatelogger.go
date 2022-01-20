package logx

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gotid/god/lib/fs"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/timex"
)

const (
	dateFormat  = "2000-01-01"
	hoursPerDay = 24

	bufferSize      = 100
	defaultDirMode  = 0o755
	defaultFileMode = 0o600
)

// ErrLogFileClosed 指示日志文件已经被关闭的错误。
var ErrLogFileClosed = errors.New("错误：日志文件已关闭")

type (
	// RotateLogger 表示一个按照指定规则滚动日志文件的记录器。
	RotateLogger struct {
		filename  string
		backup    string
		rule      RotateRule
		compress  bool
		keepDays  int
		fp        *os.File
		channel   chan []byte
		done      chan lang.PlaceholderType
		wg        sync.WaitGroup
		closeOnce sync.Once
	}

	// RotateRule 是一个用于定义日志滚动规则的接口。
	RotateRule interface {
		BackupFilename() string
		MarkRotated()
		OutdatedFiles() []string
		ShouldRotate() bool
	}

	// DailyRotateRule 是一个按天进行日志滚动的规则。
	DailyRotateRule struct {
		rotatedTime string
		filename    string
		delimiter   string
		days        int
		gzip        bool
	}
)

// DefaultRotateRule 创建一个按天滚动的默认规则。
func DefaultRotateRule(filename, delimiter string, days int, gzip bool) RotateRule {
	return &DailyRotateRule{
		rotatedTime: getNowDate(),
		filename:    filename,
		delimiter:   delimiter,
		days:        days,
		gzip:        gzip,
	}
}

// BackupFilename 返回滚动日志的文件名。
func (r *DailyRotateRule) BackupFilename() string {
	return fmt.Sprintf("%s%s%s", r.filename, r.delimiter, getNowDate())
}

// MarkRotated 标记该规则的滚动时间为当前时间。
func (r *DailyRotateRule) MarkRotated() {
	r.rotatedTime = getNowDate()
}

// OutdatedFiles 返回超过保留天数的日志文件。
func (r *DailyRotateRule) OutdatedFiles() []string {
	if r.days <= 0 {
		return nil
	}

	var pattern string
	if r.gzip {
		pattern = fmt.Sprintf("%s%s*.gz", r.filename, r.delimiter)
	} else {
		pattern = fmt.Sprintf("%s%s*", r.filename, r.delimiter)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		Errorf("获取过期日志文件失败：%s", err)
		return nil
	}

	var b strings.Builder
	boundary := time.Now().Add(-time.Hour * time.Duration(hoursPerDay*r.days)).Format(dateFormat)
	fmt.Fprintf(&b, "%s%s%s", r.filename, r.delimiter, boundary)
	if r.gzip {
		b.WriteString(".gz")
	}
	boundaryFile := b.String()

	var outdated []string
	for _, file := range files {
		// 对比文件名，判断是否为过期文件
		if file < boundaryFile {
			outdated = append(outdated, file)
		}
	}

	return outdated
}

// ShouldRotate 判断是否应该滚动日志。
func (r *DailyRotateRule) ShouldRotate() bool {
	return len(r.rotatedTime) > 0 && getNowDate() != r.rotatedTime
}

// NewLogger 返回指定文件名和滚动规则等参数的滚动日志。
func NewLogger(filename string, rule RotateRule, compress bool) (*RotateLogger, error) {
	l := &RotateLogger{
		filename: filename,
		rule:     rule,
		compress: compress,
		channel:  make(chan []byte, bufferSize),
		done:     make(chan lang.PlaceholderType),
	}
	if err := l.init(); err != nil {
		return nil, err
	}

	l.startWorker()

	return l, nil
}

// Write 将数据写入滚动日志通道。
func (l *RotateLogger) Write(data []byte) (n int, err error) {
	select {
	case l.channel <- data:
		return len(data), nil
	case <-l.done:
		log.Println(string(data))
		return 0, ErrLogFileClosed
	}
}

// Close 关闭滚动日志。
func (l *RotateLogger) Close() (err error) {
	l.closeOnce.Do(func() {
		close(l.done)
		l.wg.Wait()

		if err = l.fp.Sync(); err != nil {
			return
		}

		err = l.fp.Close()
	})

	return
}

// init 初始化滚动日志。
func (l *RotateLogger) init() error {
	l.backup = l.rule.BackupFilename()

	if _, err := os.Stat(l.filename); err != nil {
		basePath := path.Dir(l.filename)
		if _, err = os.Stat(basePath); err != nil {
			if err = os.MkdirAll(basePath, defaultDirMode); err != nil {
				return err
			}
		}

		if l.fp, err = os.Create(l.filename); err != nil {
			return err
		}
	} else if l.fp, err = os.OpenFile(l.filename, os.O_APPEND|os.O_WRONLY, defaultDirMode); err != nil {
		return err
	}

	fs.CloseOnExec(l.fp)

	return nil
}

func (l *RotateLogger) startWorker() {
	l.wg.Add(1)

	go func() {
		defer l.wg.Done()

		for {
			select {
			case event := <-l.channel:
				l.write(event)
			case <-l.done:
				return
			}
		}
	}()
}

func (l *RotateLogger) write(v []byte) {
	if l.rule.ShouldRotate() {
		if err := l.rotate(); err != nil {
			log.Println(err)
		} else {
			l.rule.MarkRotated()
		}
	}

	if l.fp != nil {
		l.fp.Write(v)
	}
}

func (l *RotateLogger) rotate() error {
	if l.fp != nil {
		err := l.fp.Close()
		l.fp = nil
		if err != nil {
			return err
		}
	}

	_, err := os.Stat(l.filename)
	if err == nil && len(l.backup) > 0 {
		backupFilename := l.getBackupFilename()
		err = os.Rename(l.filename, backupFilename)
		if err != nil {
			return err
		}

		l.postRotate(backupFilename)
	}

	l.backup = l.rule.BackupFilename()
	if l.fp, err = os.Create(l.filename); err != nil {
		fs.CloseOnExec(l.fp)
	}

	return err
}

func (l *RotateLogger) getBackupFilename() string {
	if len(l.backup) == 0 {
		return l.rule.BackupFilename()
	} else {
		return l.backup
	}
}

func (l *RotateLogger) postRotate(filename string) {
	go func() {
		// 此处不能使用 threading.GoSafe，因为logx 和 threading 会循环引用
		l.maybeCompressFile(filename)
		l.maybeDeleteOutdatedFiles()
	}()
}

func (l *RotateLogger) maybeCompressFile(filename string) {
	if !l.compress {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			ErrorStack(r)
		}
	}()

	compressLogFile(filename)
}

func (l *RotateLogger) maybeDeleteOutdatedFiles() {
	files := l.rule.OutdatedFiles()
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			Errorf("删除过期日志失败: %s", file)
		}
	}
}

func compressLogFile(filename string) {
	start := timex.Now()
	Infof("压缩日志文件: %s", filename)
	if err := fs.GzipFile(filename); err != nil {
		Errorf("压缩失败: %s", err)
	} else {
		Infof("压缩日志文件: %s, 耗时 %s", filename, timex.Since(start))
	}
}

func getNowDate() string {
	return time.Now().Format(dateFormat)
}
