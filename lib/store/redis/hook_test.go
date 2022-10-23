package redis

import (
	"context"
	red "github.com/go-redis/redis/v8"
	"github.com/gotid/god/lib/logx"
	gtrace "github.com/gotid/god/lib/trace"
	"github.com/stretchr/testify/assert"
	otrace "go.opentelemetry.io/otel/trace"
	"log"
	"strings"
	"testing"
	"time"
)

func TestHookProcessCase1(t *testing.T) {
	gtrace.StartAgent(gtrace.Config{
		Name:     "god-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer gtrace.StopAgent()

	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	ctx, err := durationHook.BeforeProcess(context.Background(), red.NewCmd(context.Background()))
	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, durationHook.AfterProcess(ctx, red.NewCmd(context.Background())))
	assert.False(t, strings.Contains(buf.String(), "slow"))
	assert.Equal(t, "redis", otrace.SpanFromContext(ctx).(interface{ Name() string }).Name())
}

func TestHookProcessCase2(t *testing.T) {
	gtrace.StartAgent(gtrace.Config{
		Name:     "god-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer gtrace.StopAgent()

	w, restore := injectLog()
	defer restore()

	ctx, err := durationHook.BeforeProcess(context.Background(), red.NewCmd(context.Background()))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "redis", otrace.SpanFromContext(ctx).(interface{ Name() string }).Name())

	time.Sleep(slowThreshold.Load() + time.Millisecond)

	assert.Nil(t, durationHook.AfterProcess(ctx, red.NewCmd(context.Background(), "foo", "bar")))
	assert.True(t, strings.Contains(w.String(), "slow"))
	assert.True(t, strings.Contains(w.String(), "trace"))
	assert.True(t, strings.Contains(w.String(), "span"))
}

func TestHookProcessCase3(t *testing.T) {
	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	assert.Nil(t, durationHook.AfterProcess(context.Background(), red.NewCmd(context.Background())))
	assert.True(t, buf.Len() == 0)
}

func TestHookProcessCase4(t *testing.T) {
	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	ctx := context.WithValue(context.Background(), startTimeKey, "foo")
	assert.Nil(t, durationHook.AfterProcess(ctx, red.NewCmd(context.Background())))
	assert.True(t, buf.Len() == 0)
}

func TestHookProcessPipelineCase1(t *testing.T) {
	gtrace.StartAgent(gtrace.Config{
		Name:     "god-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer gtrace.StopAgent()

	w, restore := injectLog()
	defer restore()

	ctx, err := durationHook.BeforeProcessPipeline(context.Background(), []red.Cmder{
		red.NewCmd(context.Background()),
	})
	assert.NoError(t, err)
	assert.Equal(t, "redis", otrace.SpanFromContext(ctx).(interface{ Name() string }).Name())

	assert.Nil(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{
		red.NewCmd(context.Background()),
	}))
	assert.False(t, strings.Contains(w.String(), "slow"))
}

func TestHookProcessPipelineCase2(t *testing.T) {
	gtrace.StartAgent(gtrace.Config{
		Name:     "god-test",
		Endpoint: "http://localhost:14268/api/traces",
		Batcher:  "jaeger",
		Sampler:  1.0,
	})
	defer gtrace.StopAgent()

	w, restore := injectLog()
	defer restore()

	ctx, err := durationHook.BeforeProcessPipeline(context.Background(), []red.Cmder{
		red.NewCmd(context.Background()),
	})
	assert.NoError(t, err)
	assert.Equal(t, "redis", otrace.SpanFromContext(ctx).(interface{ Name() string }).Name())

	time.Sleep(slowThreshold.Load() + time.Millisecond)

	assert.Nil(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{
		red.NewCmd(context.Background(), "foo", "bar"),
	}))
	assert.True(t, strings.Contains(w.String(), "slow"))
	assert.True(t, strings.Contains(w.String(), "trace"))
	assert.True(t, strings.Contains(w.String(), "span"))
}

func TestHookProcessPipelineCase3(t *testing.T) {
	w, restore := injectLog()
	defer restore()

	assert.Nil(t, durationHook.AfterProcessPipeline(context.Background(), []red.Cmder{
		red.NewCmd(context.Background()),
	}))
	assert.True(t, len(w.String()) == 0)
}

func TestHookProcessPipelineCase4(t *testing.T) {
	w, restore := injectLog()
	defer restore()

	ctx := context.WithValue(context.Background(), startTimeKey, "foo")
	assert.Nil(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{
		red.NewCmd(context.Background()),
	}))
	assert.True(t, len(w.String()) == 0)
}

func TestHookProcessPipelineCase5(t *testing.T) {
	writer := log.Writer()
	var buf strings.Builder
	log.SetOutput(&buf)
	defer log.SetOutput(writer)

	ctx := context.WithValue(context.Background(), startTimeKey, "foo")
	assert.Nil(t, durationHook.AfterProcessPipeline(ctx, []red.Cmder{red.NewCmd(context.Background())}))
	assert.True(t, buf.Len() == 0)
}

func TestLogDuration(t *testing.T) {
	w, restore := injectLog()
	defer restore()

	logDuration(context.Background(), []red.Cmder{
		red.NewCmd(context.Background(), "get", "foo"),
	}, 1*time.Second)
	assert.True(t, strings.Contains(w.String(), "get foo"))

	logDuration(context.Background(), []red.Cmder{
		red.NewCmd(context.Background(), "get", "foo"),
		red.NewCmd(context.Background(), "set", "bar", 0),
	}, 1*time.Second)
	assert.True(t, strings.Contains(w.String(), "get foo\\nset bar 0"))
}

func injectLog() (r *strings.Builder, restore func()) {
	var buf strings.Builder
	w := logx.NewWriter(&buf)
	o := logx.Reset()
	logx.SetWriter(w)

	return &buf, func() {
		logx.Reset()
		logx.SetWriter(o)
	}
}
