package metric

import (
	"github.com/gotid/god/lib/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewHistogramVec(t *testing.T) {
	hv := NewHistogramVec(&HistogramVecOpts{
		Name:    "duration_ms",
		Help:    "RPC服务器请求时长(ms)。",
		Buckets: []float64{1, 2, 3},
	})
	defer hv.close()
	hvNil := NewHistogramVec(nil)
	assert.NotNil(t, hv)
	assert.Nil(t, hvNil)
}

func TestPromHistogramVec_Observe(t *testing.T) {
	startAgent()
	histogramVec := NewHistogramVec(&HistogramVecOpts{
		Name:    "counts",
		Help:    "RPC服务器请求时长(ms)。",
		Buckets: []float64{1, 2, 3},
		Labels:  []string{"method"},
	})
	defer histogramVec.close()

	hv, _ := histogramVec.(*promHistogramVec)
	hv.Observe(2, "/Users")

	metadata := `
		# HELP counts RPC服务器请求时长(ms)。
        # TYPE counts histogram
`
	val := `
		counts_bucket{method="/Users",le="1"} 0
		counts_bucket{method="/Users",le="2"} 1
		counts_bucket{method="/Users",le="3"} 1
		counts_bucket{method="/Users",le="+Inf"} 1
		counts_sum{method="/Users"} 2
        counts_count{method="/Users"} 1
`
	err := testutil.CollectAndCompare(hv.histogram, strings.NewReader(metadata+val))
	assert.Nil(t, err)
}

func startAgent() {
	prometheus.StartAgent(prometheus.Config{
		Host: "127.0.0.1",
		Port: 9101,
		Path: "/metrics",
	})
}
