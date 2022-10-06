package stat

import (
	"github.com/gotid/god/lib/logx"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

type mockedWriter struct {
	report *StatReport
}

func (w *mockedWriter) Write(entry *StatReport) error {
	w.report = entry
	return nil
}

func TestMetrics(t *testing.T) {
	logx.Disable()
	DisableLog()
	defer logEnabled.Set(true)

	counts := []int{1, 5, 10, 100, 1000, 1000}
	for _, count := range counts {
		m := NewMetrics("foo")
		m.SetName("bar")
		for i := 0; i < count; i++ {
			m.Add(Task{
				Duration:    time.Millisecond * time.Duration(i),
				Description: strconv.Itoa(i),
			})
		}
		m.AddDrop()
		var writer mockedWriter
		SetReportWriter(&writer)
		m.executor.Flush()
		assert.Equal(t, "bar", writer.report.Name)
	}
}
