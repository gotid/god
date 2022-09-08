package logx

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gotid/god/lib/fs"
	"github.com/gotid/god/lib/lang"
)

const (
	dateFormat      = "2006-01-02"
	fileTimeFormat  = time.RFC3339
	hoursPerDay     = 24
	bufferSize      = 100
	defaultDirMode  = 0o755
	defaultFileMode = 0o600
	gzipExt         = ".gz"
	megaBytes       = 1 << 20
)

// ErrLogFileClosed 是一个表明日志文件已经关闭的错误。
var ErrLogFileClosed = errors.New("错误：日志文件已关闭")

type (
	// RotateRule 接口用于定义日志轮换规则。
	RotateRule interface {
		BackupFilename() string
		MarkRotated()
		OutdatedFiles() []string
		ShallRotate(size int64) bool
	}

	// DailyRotateRule 表示一个每天轮换日志文件的规则。
	DailyRotateRule struct {
		rotatedTime string
		filename    string
		delimiter   string
		days        int
		gzip        bool
	}

	// SizeLimitRotateRule 是一个根据大小轮换日志文件的规则。
	SizeLimitRotateRule struct {
		DailyRotateRule
		maxSize    int64
		maxBackups int
	}

	// RotateLogger 是一个可以使用给定规则轮换日志文件的记录器。
	RotateLogger struct {
		filename    string
		backup      string
		fp          *os.File
		channel     chan []byte               // 数据写入通道
		done        chan lang.PlaceholderType // 文件关闭通道
		rule        RotateRule
		compress    bool
		waitGroup   sync.WaitGroup
		closeOnce   sync.Once
		currentSize int64
	}
)

// DefaultRotateRule 是一个默认的日志轮换规则，默认为 DailyRotateRule。
func DefaultRotateRule(filename, delimiter string, days int, gzip bool) RotateRule {
	return &DailyRotateRule{
		rotatedTime: getNowDate(),
		filename:    filename,
		delimiter:   delimiter,
		days:        days,
		gzip:        gzip,
	}
}

// BackupFilename 返回基于轮换的备份文件名。
func (r *DailyRotateRule) BackupFilename() string {
	return fmt.Sprintf("%s%s%s", r.filename, r.delimiter, getNowDate())
}

// MarkRotated 标记规则的轮换时间为当前时间。
func (r *DailyRotateRule) MarkRotated() {
	r.rotatedTime = getNowDate()
}

// OutdatedFiles 返回过时的文件。
func (r *DailyRotateRule) OutdatedFiles() []string {
	if r.days <= 0 {
		return nil
	}

	var pattern string
	if r.gzip {
		pattern = fmt.Sprintf("%s%s*%s", r.filename, r.delimiter, gzipExt)
	} else {
		pattern = fmt.Sprintf("%s%s*", r.filename, r.delimiter)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		Errorf("未能删除过时的日志文件，错误：%s", err)
		return nil
	}

	var buf strings.Builder
	boundary := time.Now().Add(-time.Hour * time.Duration(hoursPerDay*r.days)).Format(dateFormat)
	fmt.Fprintf(&buf, "%s%s%s", r.filename, r.delimiter, boundary)
	if r.gzip {
		buf.WriteString(gzipExt)
	}
	boundaryFile := buf.String()

	var outdates []string
	for _, file := range files {
		if file < boundaryFile {
			outdates = append(outdates, file)
		}
	}

	return outdates
}

// ShallRotate 检查文件是否应该被轮换。
func (r *DailyRotateRule) ShallRotate(_ int64) bool {
	return len(r.rotatedTime) > 0 && getNowDate() != r.rotatedTime
}

// NewSizeLimitRotateRule 返回具有大小限制的轮换规则。
func NewSizeLimitRotateRule(filename, delimiter string, days, maxSize, maxBackups int, gzip bool) RotateRule {
	return &SizeLimitRotateRule{
		DailyRotateRule: DailyRotateRule{
			rotatedTime: getNowDateInRFC3339Format(),
			filename:    filename,
			delimiter:   delimiter,
			days:        days,
			gzip:        gzip,
		},
		maxSize:    int64(maxSize) * megaBytes,
		maxBackups: maxBackups,
	}
}

// BackupFilename 返回基于轮换的备份文件名。
func (r *SizeLimitRotateRule) BackupFilename() string {
	dir := filepath.Dir(r.filename)
	prefix, ext := r.parseFilename()
	timestamp := getNowDateInRFC3339Format()
	return filepath.Join(dir, fmt.Sprintf("%s%s%s%s", prefix, r.delimiter, timestamp, ext))
}

// MarkRotated 标记规则的轮换时间为当前时间。
func (r *SizeLimitRotateRule) MarkRotated() {
	r.rotatedTime = getNowDateInRFC3339Format()
}

// OutdatedFiles 返回过时的文件。
func (r *SizeLimitRotateRule) OutdatedFiles() []string {
	dir := filepath.Dir(r.filename)
	prefix, ext := r.parseFilename()

	var pattern string
	if r.gzip {
		pattern = fmt.Sprintf("%s%s%s%s*%s%s", dir, string(filepath.Separator),
			prefix, r.delimiter, ext, gzipExt)
	} else {
		pattern = fmt.Sprintf("%s%s%s%s*%s", dir, string(filepath.Separator),
			prefix, r.delimiter, ext)
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		Errorf("未能删除过时的日志文件，错误：%s", err)
		return nil
	}

	sort.Strings(files)

	outdated := make(map[string]lang.PlaceholderType)

	// 测试是否有太多的备份
	if r.maxBackups > 0 && len(files) > r.maxBackups {
		for _, f := range files[:len(files)-r.maxBackups] {
			outdated[f] = lang.Placeholder
		}
		files = files[len(files)-r.maxBackups:]
	}

	// 测试是否有太旧的备份
	if r.days > 0 {
		boundary := time.Now().Add(-time.Hour * time.Duration(hoursPerDay*r.days)).Format(fileTimeFormat)
		boundaryFile := filepath.Join(dir, fmt.Sprintf("%s%s%s%s", prefix, r.delimiter, boundary, ext))
		if r.gzip {
			boundaryFile += gzipExt
		}
		for _, f := range files {
			if f >= boundaryFile {
				break
			}
			outdated[f] = lang.Placeholder
		}
	}

	var result []string
	for f := range outdated {
		result = append(result, f)
	}
	return result
}

// ShallRotate 检查文件是否应该被轮换。
func (r *SizeLimitRotateRule) ShallRotate(size int64) bool {
	return r.maxSize > 0 && r.maxSize < size
}

func (r *SizeLimitRotateRule) parseFilename() (prefix, ext string) {
	logName := filepath.Base(r.filename)
	ext = filepath.Ext(r.filename)
	prefix = logName[:len(logName)-len(ext)]
	return
}

// NewLogger 返回一个给定名称和规则等参数的 RotateLogger。
func NewLogger(filename string, rule RotateRule, compress bool) (*RotateLogger, error) {
	l := &RotateLogger{
		filename: filename,
		channel:  make(chan []byte, bufferSize),
		done:     make(chan lang.PlaceholderType),
		rule:     rule,
		compress: compress,
	}
	if err := l.init(); err != nil {
		return nil, err
	}

	l.startWorker()
	return l, nil
}

// Write 将 data 写入轮换日志。
func (l *RotateLogger) Write(data []byte) (int, error) {
	select {
	case l.channel <- data:
		return len(data), nil
	case <-l.done:
		log.Println(string(data))
		return 0, ErrLogFileClosed
	}
}

// Close 关闭轮换日志记录器。
func (l *RotateLogger) Close() error {
	var err error

	l.closeOnce.Do(func() {
		close(l.done)
		l.waitGroup.Wait()

		if err = l.fp.Sync(); err != nil {
			return
		}

		err = l.fp.Close()
	})

	return err
}

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
	} else if l.fp, err = os.OpenFile(l.filename, os.O_APPEND|os.O_WRONLY, defaultFileMode); err != nil {
		return err
	}

	fs.CloseOnExec(l.fp)

	return nil
}

func (l *RotateLogger) startWorker() {
	l.waitGroup.Add(1)

	go func() {
		defer l.waitGroup.Done()

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
	if l.rule.ShallRotate(l.currentSize + int64(len(v))) {
		if err := l.rotate(); err != nil {
			log.Println(err)
		} else {
			l.rule.MarkRotated()
			l.currentSize = 0
		}
	}
	if l.fp != nil {
		l.fp.Write(v)
		l.currentSize += int64(len(v))
	}
}

func (l *RotateLogger) rotate() error {
	// 关闭文件描述符
	if l.fp != nil {
		err := l.fp.Close()
		l.fp = nil
		if err != nil {
			return err
		}
	}

	// 轮换日志文件
	_, err := os.Stat(l.filename)
	if err == nil && len(l.backup) > 0 {
		backupFilename := l.getBackupFilename()
		err = os.Rename(l.filename, backupFilename)
		if err != nil {
			return err
		}

		// 轮换后，处理压缩和删除过时文件
		l.postRotate(backupFilename)
	}

	// 记录备份文件并创建新的日志文件
	l.backup = l.rule.BackupFilename()
	if _, err = os.Create(l.filename); err == nil {
		fs.CloseOnExec(l.fp)
	}

	return err
}

func (l *RotateLogger) getBackupFilename() string {
	if len(l.backup) == 0 {
		return l.rule.BackupFilename()
	}

	return l.backup
}

func (l *RotateLogger) postRotate(file string) {
	go func() {
		l.maybeCompressFile(file)
		l.maybeDeleteOutdatedFiles()
	}()
}

func (l *RotateLogger) maybeCompressFile(file string) {
	if !l.compress {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			ErrorStack(r)
		}
	}()

	if _, err := os.Stat(file); err != nil {
		// 文件不存在或其他错误，忽略压缩
		return
	}

	compressLogFile(file)
}

func (l *RotateLogger) maybeDeleteOutdatedFiles() {
	files := l.rule.OutdatedFiles()
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			Errorf("未能删除过时文件：%s", file)
		}
	}
}

func compressLogFile(file string) {
	start := time.Now()
	Infof("压缩日志文件：%s", file)
	if err := gzipFile(file); err != nil {
		Errorf("压缩失败：%s", err)
	} else {
		Infof("压缩日志文件：%s，耗时：%s", file, time.Since(start))
	}
}

func gzipFile(file string) error {
	in, err := os.Open(file)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(fmt.Sprintf("%s%s", file, gzipExt))
	if err != nil {
		return err
	}
	defer out.Close()

	w := gzip.NewWriter(out)
	if _, err = io.Copy(w, in); err != nil {
		return err
	} else if err = w.Close(); err != nil {
		return err
	}

	return os.Remove(file)
}

func getNowDateInRFC3339Format() string {
	return time.Now().Format(fileTimeFormat)
}

func getNowDate() string {
	return time.Now().Format(dateFormat)
}
