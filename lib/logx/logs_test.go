package logx

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"encoding/json"
	"github.com/stretchr/testify/assert"
)

type mockWriter struct {
	lock    sync.Mutex
	builder strings.Builder
}

func (mw *mockWriter) Close() error {
	return nil
}

func (mw *mockWriter) Debug(v any, fields ...LogField) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelDebug, v, fields...)
}

func (mw *mockWriter) Info(v any, fields ...LogField) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelInfo, v, fields...)
}

func (mw *mockWriter) Alert(v any) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelAlert, v)
}

func (mw *mockWriter) Error(v any, fields ...LogField) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelError, v, fields...)
}

func (mw *mockWriter) Severe(v any) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelSevere, v)
}

func (mw *mockWriter) Slow(v any, fields ...LogField) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelSlow, v, fields...)
}

func (mw *mockWriter) Stack(v any) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelError, v)
}

func (mw *mockWriter) Stat(v any, fields ...LogField) {
	mw.lock.Lock()
	defer mw.lock.Unlock()
	output(&mw.builder, levelStat, v, fields...)
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

type ValStringer struct {
	val string
}

func (v ValStringer) String() string {
	return v.val
}

func TestField(t *testing.T) {
	tests := []struct {
		name string
		f    LogField
		want map[string]any
	}{
		{
			name: "error",
			f:    Field("foo", errors.New("bar")),
			want: map[string]any{

				"foo": "bar",
			},
		},
		{
			name: "errors",
			f:    Field("foo", []error{errors.New("bar"), errors.New("baz")}),
			want: map[string]any{
				"foo": []any{"bar", "baz"},
			},
		},
		{
			name: "strings",
			f:    Field("foo", []string{"bar", "baz"}),
			want: map[string]any{
				"foo": []any{"bar", "baz"},
			},
		},
		{
			name: "duration",
			f:    Field("foo", time.Second),
			want: map[string]any{
				"foo": "1s",
			},
		},
		{
			name: "durations",
			f:    Field("foo", []time.Duration{time.Second, 2 * time.Second}),
			want: map[string]any{
				"foo": []any{"1s", "2s"},
			},
		},
		{
			name: "times",
			f: Field("foo", []time.Time{
				time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC),
			}),
			want: map[string]any{
				"foo": []any{"2020-01-01 00:00:00 +0000 UTC", "2020-01-02 00:00:00 +0000 UTC"},
			},
		},
		{
			name: "stringer",
			f:    Field("foo", ValStringer{val: "bar"}),
			want: map[string]any{
				"foo": "bar",
			},
		},
		{
			name: "stringers",
			f:    Field("foo", []fmt.Stringer{ValStringer{val: "bar"}, ValStringer{val: "baz"}}),
			want: map[string]any{
				"foo": []any{"bar", "baz"},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			w := new(mockWriter)
			old := writer.Swap(w)
			defer writer.Store(old)

			Infow("foo", test.f)
			validateFields(t, w.String(), test.want)
		})
	}
}

func TestFileLineFileMode(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	file, line := getFileLine()
	Error("anything")
	assert.True(t, w.Contains(fmt.Sprintf("%s:%d", file, line+1)))

	file, line = getFileLine()
	Errorf("anything %s", "format")
	assert.True(t, w.Contains(fmt.Sprintf("%s:%d", file, line+1)))
}

func TestFileLineConsoleMode(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	file, line := getFileLine()
	Error("anything")
	assert.True(t, w.Contains(fmt.Sprintf("%s:%d", file, line+1)))

	w.Reset()
	file, line = getFileLine()
	Errorf("anything %s", "format")
	assert.True(t, w.Contains(fmt.Sprintf("%s:%d", file, line+1)))
}

func TestStructuredLogAlert(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelAlert, w, func(v ...any) {
		Alert(fmt.Sprint(v...))
	})
}

func TestStructuredLogError(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelError, w, func(v ...any) {
		Error(fmt.Sprint(v...))
	})
}

func TestStructuredLogErrorf(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelError, w, func(v ...any) {
		Errorf("%s", fmt.Sprint(v...))
	})
}

func TestStructuredLogErrorv(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelError, w, func(v ...any) {
		Errorv(fmt.Sprint(v...))
	})
}

func TestStructuredLogErrorw(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelError, w, func(v ...any) {
		Errorw(fmt.Sprint(v...), Field("foo", "bar"))
	})
}

func TestStructuredLogInfo(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelInfo, w, func(v ...any) {
		Info(v...)
	})
}

func TestStructuredLogInfof(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelInfo, w, func(v ...any) {
		Infof("%s", fmt.Sprint(v...))
	})
}

func TestStructuredLogInfov(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelInfo, w, func(v ...any) {
		Infov(fmt.Sprint(v...))
	})
}

func TestStructuredLogInfow(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelInfo, w, func(v ...any) {
		Infow(fmt.Sprint(v...), Field("foo", "bar"))
	})
}

func TestStructuredLogInfoConsoleAny(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLogConsole(t, w, func(v ...any) {
		old := atomic.LoadUint32(&encoding)
		atomic.StoreUint32(&encoding, plainEncodingType)
		defer func() {
			atomic.StoreUint32(&encoding, old)
		}()

		Infov(v)
	})
}

func TestStructuredLogInfoConsoleAnyString(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLogConsole(t, w, func(v ...any) {
		old := atomic.LoadUint32(&encoding)
		atomic.StoreUint32(&encoding, plainEncodingType)
		defer func() {
			atomic.StoreUint32(&encoding, old)
		}()

		Infov(fmt.Sprint(v...))
	})
}

func TestStructuredLogInfoConsoleAnyError(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLogConsole(t, w, func(v ...any) {
		old := atomic.LoadUint32(&encoding)
		atomic.StoreUint32(&encoding, plainEncodingType)
		defer func() {
			atomic.StoreUint32(&encoding, old)
		}()

		Infov(errors.New(fmt.Sprint(v...)))
	})
}

func TestStructuredLogInfoConsoleAnyStringer(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLogConsole(t, w, func(v ...any) {
		old := atomic.LoadUint32(&encoding)
		atomic.StoreUint32(&encoding, plainEncodingType)
		defer func() {
			atomic.StoreUint32(&encoding, old)
		}()

		Infov(ValStringer{
			val: fmt.Sprint(v...),
		})
	})
}

func TestStructuredLogInfoConsoleText(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLogConsole(t, w, func(v ...any) {
		old := atomic.LoadUint32(&encoding)
		atomic.StoreUint32(&encoding, plainEncodingType)
		defer func() {
			atomic.StoreUint32(&encoding, old)
		}()

		Info(fmt.Sprint(v...))
	})
}

func TestStructuredLogSlow(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelSlow, w, func(v ...any) {
		Slow(v...)
	})
}

func TestStructuredLogSlowf(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelSlow, w, func(v ...any) {
		Slowf(fmt.Sprint(v...))
	})
}

func TestStructuredLogSlowv(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelSlow, w, func(v ...any) {
		Slowv(fmt.Sprint(v...))
	})
}

func TestStructuredLogSloww(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelSlow, w, func(v ...any) {
		Sloww(fmt.Sprint(v...), Field("foo", time.Second))
	})
}

func TestStructuredLogStat(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelStat, w, func(v ...any) {
		Stat(v...)
	})
}

func TestStructuredLogStatf(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelStat, w, func(v ...any) {
		Statf(fmt.Sprint(v...))
	})
}

func TestStructuredLogSevere(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelSevere, w, func(v ...any) {
		Severe(v...)
	})
}

func TestStructuredLogSeveref(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	doTestStructuredLog(t, levelSevere, w, func(v ...any) {
		Severef(fmt.Sprint(v...))
	})
}

func TestStructuredLogWithDuration(t *testing.T) {
	const message = "hellox there"
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Info(message)
	var entry logEntry
	if err := json.Unmarshal([]byte(w.String()), &entry); err != nil {
		t.Error(err)
	}
	assert.Equal(t, levelInfo, entry[levelKey])
	assert.Equal(t, message, entry[contentKey])
	assert.Equal(t, "1000.0ms", entry[durationKey])
}

func TestSetLevel(t *testing.T) {
	SetLevel(ErrorLevel)
	const message = "hellox there"
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	Info(message)
	assert.Equal(t, 0, w.builder.Len())
}

func TestSetLevelTwiceWithMode(t *testing.T) {
	testModes := []string{
		"mode",
		"console",
		"volume",
	}
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	for _, mode := range testModes {
		testSetLevelTwiceWithMode(t, mode, w)
	}
}

func TestSetLevelWithDuration(t *testing.T) {
	SetLevel(ErrorLevel)
	const message = "hellox there"
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Info(message)
	assert.Equal(t, 0, w.builder.Len())
}

func TestErrorfWithWrappedError(t *testing.T) {
	SetLevel(ErrorLevel)
	const message = "there"
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	Errorf("hellox %w", errors.New(message))
	assert.True(t, strings.Contains(w.String(), "hellox there"))
}

func TestMustNil(t *testing.T) {
	Must(nil)
}

func TestSetup(t *testing.T) {
	defer func() {
		SetLevel(InfoLevel)
		atomic.StoreUint32(&encoding, jsonEncodingType)
	}()

	MustSetup(Config{
		ServiceName: "any",
		Mode:        "console",
	})
	MustSetup(Config{
		ServiceName: "any",
		Mode:        "file",
		Path:        os.TempDir(),
	})
	MustSetup(Config{
		ServiceName: "any",
		Mode:        "volume",
		Path:        os.TempDir(),
	})
	MustSetup(Config{
		ServiceName: "any",
		Mode:        "console",
		TimeFormat:  timeFormat,
	})
	MustSetup(Config{
		ServiceName: "any",
		Mode:        "console",
		Encoding:    plainEncoding,
	})

	assert.NotNil(t, setupWithVolume(Config{}))
	assert.NotNil(t, setupWithFiles(Config{}))
	assert.Nil(t, setupWithFiles(Config{
		ServiceName: "any",
		Path:        os.TempDir(),
		Compress:    true,
		KeepDays:    1,
	}))
	setupLogLevel(Config{
		Level: levelInfo,
	})
	setupLogLevel(Config{
		Level: levelError,
	})
	setupLogLevel(Config{
		Level: levelSevere,
	})
	_, err := createOutput("")
	assert.NotNil(t, err)
	Disable()
	SetLevel(InfoLevel)
	atomic.StoreUint32(&encoding, jsonEncodingType)
}

func TestDisable(t *testing.T) {
	Disable()

	var opt logOptions
	WithKeepDays(1)(&opt)
	WithGzip()(&opt)
	assert.Nil(t, Close())
	assert.Nil(t, Close())
}

func TestDisableStat(t *testing.T) {
	DisableStat()

	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)
	Stat("这是统计信息，但被禁用了...")
	assert.Equal(t, 0, w.builder.Len())
	Info("这是通知信息")
	assert.True(t, strings.Contains(w.String(), "这是通知信息"), w.String())
}

func TestSetWriter(t *testing.T) {
	atomic.StoreUint32(&disableLog, 0)
	Reset()
	SetWriter(nopWriter{})
	assert.NotNil(t, writer.Load())
	assert.True(t, writer.Load() == nopWriter{})
	mocked := new(mockWriter)
	SetWriter(mocked)
	assert.Equal(t, mocked, writer.Load())
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

var (
	s           = []byte("Sending #11 notification (id: 1451875113812010473) in #1 connection")
	pool        = make(chan []byte, 1)
	_    Writer = (*mockWriter)(nil)
)

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

func put(b []byte) {
	select {
	case pool <- b:
	default:
	}
}

func doTestStructuredLog(t *testing.T, level string, w *mockWriter, write func(...any)) {
	const message = "hellox there"
	write(message)
	var entry logEntry
	if err := json.Unmarshal([]byte(w.String()), &entry); err != nil {
		t.Error(err)
	}
	assert.Equal(t, level, entry[levelKey])
	val, ok := entry[contentKey]
	assert.True(t, ok)
	assert.True(t, strings.Contains(val.(string), message))
}

func doTestStructuredLogConsole(t *testing.T, w *mockWriter, write func(...any)) {
	const message = "hellox there"
	write(message)
	assert.True(t, strings.Contains(w.String(), message))
}

func testSetLevelTwiceWithMode(t *testing.T, mode string, w *mockWriter) {
	writer.Store(nil)
	Setup(Config{
		Mode:  mode,
		Level: "error",
		Path:  "/dev/null",
	})
	Setup(Config{
		Mode:  mode,
		Level: "info",
		Path:  "/dev/null",
	})
	const message = "hellox there"
	Info(message)
	assert.Equal(t, 0, w.builder.Len())
	Infof(message)
	assert.Equal(t, 0, w.builder.Len())
	ErrorStack(message)
	assert.Equal(t, 0, w.builder.Len())
	ErrorStackf(message)
	assert.Equal(t, 0, w.builder.Len())
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

func validateFields(t *testing.T, content string, fields map[string]any) {
	var m map[string]any
	if err := json.Unmarshal([]byte(content), &m); err != nil {
		t.Error(err)
	}

	for k, v := range fields {
		if reflect.TypeOf(v).Kind() == reflect.Slice {
			assert.EqualValues(t, v, m[k])
		} else {
			assert.Equal(t, v, m[k], content)
		}
	}
}
