package logx

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/stretchr/testify/assert"
)

func TestDurationLogger_Error(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Error("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Errorf(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Errorf("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Errorv(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Errorv("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Errorw(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Errorw("foo", Field("foo", "bar"))
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
	assert.True(t, strings.Contains(w.String(), "foo"), w.String())
	assert.True(t, strings.Contains(w.String(), "bar"), w.String())
}

func TestDurationLogger_Info(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Info("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_InfoConsole(t *testing.T) {
	old := atomic.LoadUint32(&encoding)
	atomic.StoreUint32(&encoding, plainEncodingType)
	defer atomic.StoreUint32(&encoding, old)

	w := new(mockWriter)
	o := writer.Swap(w)
	defer writer.Store(o)

	WithDuration(time.Second).Info("foo")
	assert.True(t, strings.Contains(w.String(), "ms"), w.String())
}

func TestDurationLogger_Infof(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Infof("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Infov(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Infov("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Infow(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Infow("foo", Field("foo", "bar"))
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
	assert.True(t, strings.Contains(w.String(), "foo"), w.String())
	assert.True(t, strings.Contains(w.String(), "bar"), w.String())
}

func TestDurationLogger_WithContextInfow(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, _ := tp.Tracer("foo").Start(context.Background(), "bar")

	WithDuration(time.Second).WithContext(ctx).Infow("foo", Field("foo", "bar"))
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
	assert.True(t, strings.Contains(w.String(), "foo"), w.String())
	assert.True(t, strings.Contains(w.String(), "bar"), w.String())
	assert.True(t, strings.Contains(w.String(), "trace"), w.String())
	assert.True(t, strings.Contains(w.String(), "span"), w.String())
}

func TestDurationLogger_Slow(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).Slow("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Slowf(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).WithDuration(time.Hour).Slowf("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Slowv(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).WithDuration(time.Hour).Slowv("foo")
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
}

func TestDurationLogger_Sloww(t *testing.T) {
	w := new(mockWriter)
	old := writer.Swap(w)
	defer writer.Store(old)

	WithDuration(time.Second).WithDuration(time.Hour).Sloww("foo", Field("foo", "bar"))
	assert.True(t, strings.Contains(w.String(), "duration"), w.String())
	assert.True(t, strings.Contains(w.String(), "foo"), w.String())
	assert.True(t, strings.Contains(w.String(), "bar"), w.String())
}
