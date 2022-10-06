package logx

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddGlobalFields(t *testing.T) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)
	old := Reset()
	SetWriter(writer)
	defer SetWriter(old)

	Info("hello")
	buf.Reset()

	AddGlobalFields(Field("a", "1"), Field("b", "2"))
	AddGlobalFields(Field("c", "3"))
	Info("world")

	var m map[string]interface{}
	assert.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	assert.Equal(t, "1", m["a"])
	assert.Equal(t, "2", m["b"])
	assert.Equal(t, "3", m["c"])
}

func TestWithFields(t *testing.T) {
	ctx := ContextWithFields(context.Background(), Field("a", 1), Field("b", 2))
	vals := ctx.Value(fieldsContextKey)
	assert.NotNil(t, vals)
	fields, ok := vals.([]LogField)
	assert.True(t, ok)
	assert.EqualValues(t, []LogField{Field("a", 1), Field("b", 2)}, fields)
}

func TestWithFieldsAppend(t *testing.T) {
	var dummyKey struct{}
	ctx := context.WithValue(context.Background(), dummyKey, "dummy")
	ctx = ContextWithFields(ctx, Field("a", 1), Field("b", 2))
	ctx = ContextWithFields(ctx, Field("c", 3), Field("d", 4))
	vals := ctx.Value(fieldsContextKey)
	assert.NotNil(t, vals)
	fields, ok := vals.([]LogField)
	assert.True(t, ok)
	assert.Equal(t, "dummy", ctx.Value(dummyKey))
	assert.EqualValues(t, []LogField{
		Field("a", 1),
		Field("b", 2),
		Field("c", 3),
		Field("d", 4),
	}, fields)
}

func BenchmarkAtomicValue(b *testing.B) {
	b.ReportAllocs()

	var container atomic.Value
	fields := []LogField{
		Field("a", "b"),
		Field("c", "d"),
		Field("e", "f"),
	}
	container.Store(&fields)

	for i := 0; i < b.N; i++ {
		vs := container.Load()
		if vs != nil {
			_ = *vs.(*[]LogField)
		}
	}
}

func BenchmarkRWMutex(b *testing.B) {
	b.ReportAllocs()

	var lock sync.RWMutex
	fields := []LogField{
		Field("a", "b"),
		Field("c", "d"),
		Field("e", "f"),
	}

	for i := 0; i < b.N; i++ {
		lock.RLock()
		_ = fields
		lock.RUnlock()
	}
}
