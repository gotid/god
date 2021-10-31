package logx

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel"
)

const (
	traceKey = "trace"
	spanKey  = "span"

	testLog = "Stay hungry, stay foolish."
)

func TestTraceLog(t *testing.T) {
	var buf mockWriter
	atomic.SwapUint32(&initialized, 1)

	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, _ := tp.Tracer("foo").Start(context.Background(), "bar")
	WithContext(ctx).(*traceLogger).write(&buf, levelInfo, testLog)
	assert.True(t, strings.Contains(buf.String(), traceKey))
	assert.True(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())
}

func TestTraceError(t *testing.T) {
	var buf mockWriter
	atomic.StoreUint32(&initialized, 1)
	errorLog = newLogWriter(log.New(&buf, "", flags))
	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, _ := tp.Tracer("foo").Start(context.Background(), "bar")
	l := WithContext(ctx).(*traceLogger)
	SetLevel(InfoLevel)

	l.WithDuration(time.Second).Error(testLog)
	assert.True(t, strings.Contains(buf.String(), traceKey))
	assert.True(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())

	buf.Reset()
	l.WithDuration(time.Second).Errorf(testLog)
	assert.True(t, strings.Contains(buf.String(), traceKey))
	assert.True(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())
}

func TestTraceInfo(t *testing.T) {
	var buf mockWriter
	atomic.StoreUint32(&initialized, 1)
	infoLog = newLogWriter(log.New(&buf, "", flags))
	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, _ := tp.Tracer("foo").Start(context.Background(), "bar")
	l := WithContext(ctx).(*traceLogger)
	SetLevel(InfoLevel)

	l.WithDuration(time.Second).Info(testLog)
	assert.True(t, strings.Contains(buf.String(), traceKey))
	assert.True(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())

	buf.Reset()
	l.WithDuration(time.Second).Infof(testLog)
	assert.True(t, strings.Contains(buf.String(), traceKey))
	assert.True(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())
}

func TestTraceSlow(t *testing.T) {
	var buf mockWriter
	atomic.StoreUint32(&initialized, 1)
	slowLog = newLogWriter(log.New(&buf, "", flags))
	otp := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(otp)

	ctx, _ := tp.Tracer("foo").Start(context.Background(), "bar")
	l := WithContext(ctx).(*traceLogger)
	SetLevel(InfoLevel)

	l.WithDuration(time.Second).Slow(testLog)
	assert.True(t, strings.Contains(buf.String(), traceKey))
	assert.True(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())

	buf.Reset()
	l.WithDuration(time.Second).Slowf(testLog)
	assert.True(t, strings.Contains(buf.String(), traceKey))
	assert.True(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())
}

func TestTraceWithoutContext(t *testing.T) {
	var buf mockWriter
	atomic.StoreUint32(&initialized, 1)
	infoLog = newLogWriter(log.New(&buf, "", flags))
	l := WithContext(context.Background()).(*traceLogger)
	SetLevel(InfoLevel)

	l.WithDuration(time.Second).Info(testLog)
	assert.False(t, strings.Contains(buf.String(), traceKey))
	assert.False(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())

	buf.Reset()
	l.WithDuration(time.Second).Infof(testLog)
	assert.False(t, strings.Contains(buf.String(), traceKey))
	assert.False(t, strings.Contains(buf.String(), spanKey))
	fmt.Println(buf.String())
}
