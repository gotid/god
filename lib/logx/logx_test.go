package logx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	s    = []byte("Sending #11 notification (id: 1451875113812010473) in #1 connection")
	pool = make(chan []byte, 1)
)

type mockWriter struct {
	lock    sync.Mutex
	builder strings.Builder
}

func (mw *mockWriter) Write(data []byte) (int, error) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	return mw.builder.Write(data)
}

func (mw *mockWriter) Close() error {
	return nil
}

func (mw *mockWriter) Contains(text string) bool {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	return strings.Contains(mw.builder.String(), text)
}

func (mw *mockWriter) Reset() {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	mw.builder.Reset()
}

func (mw *mockWriter) String() string {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	return mw.builder.String()
}

func TestErrorfWithWrappedError(t *testing.T) {
	SetLevel(ErrorLevel)
	const message = "there"
	writer := new(mockWriter)
	errorLog = writer
	atomic.StoreUint32(&initialized, 1)
	Errorf("hello %w", errors.New(message))
	assert.True(t, strings.Contains(writer.builder.String(), "hello there"))
}

func TestFileLineFileMode(t *testing.T) {
	writer := new(mockWriter)
	errorLog = writer
	atomic.StoreUint32(&initialized, 1)
	file, line := getFileLine()
	Error("anything")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))

	writer.Reset()
	file, line = getFileLine()
	Errorf("anything %s", "format")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))
}

func TestFileLineConsoleMode(t *testing.T) {
	writer := new(mockWriter)
	writeConsole = true
	errorLog = newLogWriter(log.New(writer, "[ERROR] ", flags))
	atomic.StoreUint32(&initialized, 1)
	file, line := getFileLine()
	Error("anything")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))

	writer.Reset()
	file, line = getFileLine()
	Errorf("anything %s", "format")
	assert.True(t, writer.Contains(fmt.Sprintf("%s:%d", file, line+1)))
}

func TestStructedLogAlert(t *testing.T) {
	doTestStructedLog(t, alertLevel, func(writer io.WriteCloser) {
		errorLog = writer
	}, func(v ...interface{}) {
		Alert(fmt.Sprint(v...))
	})
}

func TestStructedLogError(t *testing.T) {
	doTestStructedLog(t, errorLevel, func(writer io.WriteCloser) {
		errorLog = writer
	}, func(v ...interface{}) {
		Error(v...)
	})
}

func TestStructedLogErrorf(t *testing.T) {
	doTestStructedLog(t, errorLevel, func(writer io.WriteCloser) {
		errorLog = writer
	}, func(v ...interface{}) {
		Errorf("%s", fmt.Sprint(v...))
	})
}

func TestStructedLogErrorv(t *testing.T) {
	doTestStructedLog(t, errorLevel, func(writer io.WriteCloser) {
		errorLog = writer
	}, func(v ...interface{}) {
		Errorv(fmt.Sprint(v...))
	})
}

func TestStructedLogInfo(t *testing.T) {
	doTestStructedLog(t, infoLevel, func(writer io.WriteCloser) {
		infoLog = writer
	}, func(v ...interface{}) {
		Info(v...)
	})
}

func TestStructedLogInfof(t *testing.T) {
	doTestStructedLog(t, infoLevel, func(writer io.WriteCloser) {
		infoLog = writer
	}, func(v ...interface{}) {
		Infof("%s", fmt.Sprint(v...))
	})
}

func TestStructedLogInfov(t *testing.T) {
	doTestStructedLog(t, infoLevel, func(writer io.WriteCloser) {
		infoLog = writer
	}, func(v ...interface{}) {
		Infov(fmt.Sprint(v...))
	})
}

func TestStructedLogSlow(t *testing.T) {
	doTestStructedLog(t, slowLevel, func(writer io.WriteCloser) {
		slowLog = writer
	}, func(v ...interface{}) {
		Slow(v...)
	})
}

func TestStructedLogSlowf(t *testing.T) {
	doTestStructedLog(t, slowLevel, func(writer io.WriteCloser) {
		slowLog = writer
	}, func(v ...interface{}) {
		Slowf(fmt.Sprint(v...))
	})
}

func TestStructedLogSlowv(t *testing.T) {
	doTestStructedLog(t, slowLevel, func(writer io.WriteCloser) {
		slowLog = writer
	}, func(v ...interface{}) {
		Slowv(fmt.Sprint(v...))
	})
}

func TestStructedLogStat(t *testing.T) {
	doTestStructedLog(t, statLevel, func(writer io.WriteCloser) {
		statLog = writer
	}, func(v ...interface{}) {
		Stat(v...)
	})
}

func TestStructedLogStatf(t *testing.T) {
	doTestStructedLog(t, statLevel, func(writer io.WriteCloser) {
		statLog = writer
	}, func(v ...interface{}) {
		Statf(fmt.Sprint(v...))
	})
}

func TestStructedLogSevere(t *testing.T) {
	doTestStructedLog(t, serverLevel, func(writer io.WriteCloser) {
		severeLog = writer
	}, func(v ...interface{}) {
		Severe(v...)
	})
}

func TestStructedLogSeveref(t *testing.T) {
	doTestStructedLog(t, serverLevel, func(writer io.WriteCloser) {
		severeLog = writer
	}, func(v ...interface{}) {
		Severef(fmt.Sprint(v...))
	})
}

func TestStructedLogWithDuration(t *testing.T) {
	const message = "hello there"
	writer := new(mockWriter)
	infoLog = writer
	atomic.StoreUint32(&initialized, 1)
	WithDuration(time.Second).Info(message)
	var entry logEntry
	if err := json.Unmarshal([]byte(writer.builder.String()), &entry); err != nil {
		t.Error(err)
	}
	assert.Equal(t, infoLevel, entry.Level)
	assert.Equal(t, message, entry.Content)
	assert.Equal(t, "1000.0ms", entry.Duration)
}

func TestSetLevel(t *testing.T) {
	SetLevel(ErrorLevel)
	const message = "hello there"
	writer := new(mockWriter)
	infoLog = writer
	atomic.StoreUint32(&initialized, 1)
	Info(message)
	assert.Equal(t, 0, writer.builder.Len())
}

func TestSetLevelTwiceWithMode(t *testing.T) {
	testModes := []string{
		"mode",
		"console",
		"volumn",
	}
	for _, mode := range testModes {
		testSetLevelTwiceWithMode(t, mode)
	}
}

func TestSetLevelWithDuration(t *testing.T) {
	SetLevel(ErrorLevel)
	const message = "hello there"
	writer := new(mockWriter)
	infoLog = writer
	atomic.StoreUint32(&initialized, 1)
	WithDuration(time.Second).Info(message)
	assert.Equal(t, 0, writer.builder.Len())
}

func TestMustNil(t *testing.T) {
	Must(nil)
}

func TestSetup(t *testing.T) {
	MustSetup(LogConf{
		ServiceName: "any",
		Mode:        "console",
	})
	MustSetup(LogConf{
		ServiceName: "any",
		Mode:        "file",
		Path:        os.TempDir(),
	})
	MustSetup(LogConf{
		ServiceName: "any",
		Mode:        "volume",
		Path:        os.TempDir(),
	})
	assert.NotNil(t, setupWithVolume(LogConf{}))
	assert.NotNil(t, setupWithFiles(LogConf{}))
	assert.Nil(t, setupWithFiles(LogConf{
		ServiceName: "any",
		Path:        os.TempDir(),
		Compress:    true,
		KeepDays:    1,
	}))
	setupLogLevel(LogConf{
		Level: infoLevel,
	})
	setupLogLevel(LogConf{
		Level: errorLevel,
	})
	setupLogLevel(LogConf{
		Level: serverLevel,
	})
	_, err := createOutput("")
	assert.NotNil(t, err)
	Disable()
}

func TestDisable(t *testing.T) {
	Disable()

	var opt logOptions
	WithKeepDays(1)(&opt)
	WithGzip()(&opt)
	assert.Nil(t, Close())
	writeConsole = false
	assert.Nil(t, Close())
}

func TestDisableStat(t *testing.T) {
	DisableStat()

	const message = "hello there"
	writer := new(mockWriter)
	statLog = writer
	atomic.StoreUint32(&initialized, 1)
	Stat(message)
	assert.Equal(t, 0, writer.builder.Len())
}

func TestWithGzip(t *testing.T) {
	fn := WithGzip()
	var opt logOptions
	fn(&opt)
	assert.True(t, opt.gzipEnabled)
}

func TestWithKeepDays(t *testing.T) {
	fn := WithKeepDays(1)
	var opt logOptions
	fn(&opt)
	assert.Equal(t, 1, opt.keepDays)
}

func BenchmarkCopyByteSliceAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf []byte
		buf = append(buf, getTimestamp()...)
		buf = append(buf, ' ')
		buf = append(buf, s...)
		_ = buf
	}
}

func BenchmarkCopyByteSliceAllocExactly(b *testing.B) {
	for i := 0; i < b.N; i++ {
		now := []byte(getTimestamp())
		buf := make([]byte, len(now)+1+len(s))
		n := copy(buf, now)
		buf[n] = ' '
		copy(buf[n+1:], s)
	}
}

func BenchmarkCopyByteSlice(b *testing.B) {
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = make([]byte, len(s))
		copy(buf, s)
	}
	fmt.Fprint(ioutil.Discard, buf)
}

func BenchmarkCopyOnWriteByteSlice(b *testing.B) {
	var buf []byte
	for i := 0; i < b.N; i++ {
		size := len(s)
		buf = s[:size:size]
	}
	fmt.Fprint(ioutil.Discard, buf)
}

func BenchmarkCacheByteSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dup := fetch()
		copy(dup, s)
		put(dup)
	}
}

func BenchmarkLogs(b *testing.B) {
	b.ReportAllocs()

	log.SetOutput(ioutil.Discard)
	for i := 0; i < b.N; i++ {
		Info(i)
	}
}

func fetch() []byte {
	select {
	case b := <-pool:
		return b
	default:
	}
	return make([]byte, 4096)
}

func getFileLine() (string, int) {
	_, file, line, _ := runtime.Caller(1)
	short := file

	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}

	return short, line
}

func put(b []byte) {
	select {
	case pool <- b:
	default:
	}
}

func doTestStructedLog(t *testing.T, level string, setup func(writer io.WriteCloser),
	write func(...interface{})) {
	const message = "hello there"
	writer := new(mockWriter)
	setup(writer)
	atomic.StoreUint32(&initialized, 1)
	write(message)
	var entry logEntry
	if err := json.Unmarshal([]byte(writer.builder.String()), &entry); err != nil {
		t.Error(err)
	}
	assert.Equal(t, level, entry.Level)
	val, ok := entry.Content.(string)
	assert.True(t, ok)
	assert.True(t, strings.Contains(val, message))
}

func testSetLevelTwiceWithMode(t *testing.T, mode string) {
	Setup(LogConf{
		Mode:  mode,
		Level: "error",
		Path:  "/dev/null",
	})
	Setup(LogConf{
		Mode:  mode,
		Level: "info",
		Path:  "/dev/null",
	})
	const message = "hello there"
	writer := new(mockWriter)
	infoLog = writer
	atomic.StoreUint32(&initialized, 1)
	Info(message)
	assert.Equal(t, 0, writer.builder.Len())
	Infof(message)
	assert.Equal(t, 0, writer.builder.Len())
	ErrorStack(message)
	assert.Equal(t, 0, writer.builder.Len())
	ErrorStackf(message)
	assert.Equal(t, 0, writer.builder.Len())
}
