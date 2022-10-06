package metric

import (
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewGaugeVec(t *testing.T) {
	gaugeVec := NewGaugeVec(&GaugeVecOpts{
		Namespace: "rpc_server",
		Subsystem: "requests",
		Name:      "duration",
		Help:      "RPC服务器请求时长（毫秒）。",
	})
	defer gaugeVec.close()
	gaugeVecNil := NewGaugeVec(nil)
	assert.NotNil(t, gaugeVec)
	assert.Nil(t, gaugeVecNil)
}

func TestGaugeInc(t *testing.T) {
	startAgent()
	gaugeVec := NewGaugeVec(&GaugeVecOpts{
		Namespace: "rpc_client2",
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "RPC服务器请求时长（毫秒）。",
		Labels:    []string{"path"},
	})
	defer gaugeVec.close()
	gv, _ := gaugeVec.(*promGaugeVec)
	gv.Inc("/users")
	gv.Inc("/users")
	r := testutil.ToFloat64(gv.gauge)
	assert.Equal(t, float64(2), r)
}

func TestGaugeAdd(t *testing.T) {
	startAgent()
	gaugeVec := NewGaugeVec(&GaugeVecOpts{
		Namespace: "rpc_client",
		Subsystem: "request",
		Name:      "duration_ms",
		Help:      "RPC服务器请求时长（毫秒）。",
		Labels:    []string{"path"},
	})
	defer gaugeVec.close()
	gv, _ := gaugeVec.(*promGaugeVec)
	gv.Add(-10, "/classroom")
	gv.Add(30, "/classroom")
	r := testutil.ToFloat64(gv.gauge)
	assert.Equal(t, float64(20), r)
}

func TestGaugeSet(t *testing.T) {
	startAgent()
	gaugeVec := NewGaugeVec(&GaugeVecOpts{
		Namespace: "http_client",
		Subsystem: "request",
		Name:      "duration_ms",
		Help:      "HTTP客户端请求时长（毫秒）。",
		Labels:    []string{"path"},
	})
	gaugeVec.close()
	gv, _ := gaugeVec.(*promGaugeVec)
	gv.Set(666, "/users")
	r := testutil.ToFloat64(gv.gauge)
	assert.Equal(t, float64(666), r)
}
