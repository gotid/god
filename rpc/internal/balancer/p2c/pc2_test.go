package p2c

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"testing"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mathx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
)

func init() {
	logx.Disable()
}

func TestPicker_PickNil(t *testing.T) {
	builder := new(pickerBuilder)
	picker := builder.Build(base.PickerBuildInfo{})
	_, err := picker.Pick(balancer.PickInfo{
		FullMethodName: "/",
		Ctx:            context.Background(),
	})
	assert.NotNil(t, err)
	fmt.Println(err)
}

func TestPicker_Pick(t *testing.T) {
	tests := []struct {
		name       string
		candidates int
		threshold  float64
	}{
		{
			name:       "单个",
			candidates: 1,
			threshold:  0.9,
		},
		{
			name:       "两个",
			candidates: 2,
			threshold:  0.5,
		},
		{
			name:       "多个",
			candidates: 100,
			threshold:  0.95,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			const total = 10000
			builder := new(pickerBuilder)
			ready := make(map[balancer.SubConn]base.SubConnInfo)
			for i := 0; i < test.candidates; i++ {
				ready[new(mockClientConn)] = base.SubConnInfo{
					Address: resolver.Address{
						Addr: strconv.Itoa(i),
					},
				}
			}

			picker := builder.Build(base.PickerBuildInfo{
				ReadySCs: ready,
			})
			var wg sync.WaitGroup
			wg.Add(total)
			for i := 0; i < total; i++ {
				result, err := picker.Pick(balancer.PickInfo{
					FullMethodName: "/",
					Ctx:            context.Background(),
				})
				assert.Nil(t, err)
				if i%100 == 0 {
					err = status.Error(codes.DeadlineExceeded, "超时啦")
				}
				go func() {
					runtime.Gosched()
					result.Done(balancer.DoneInfo{
						Err: err,
					})
					wg.Done()
				}()
			}

			wg.Wait()
			dist := make(map[interface{}]int)
			conns := picker.(*p2cPicker).conns
			for _, conn := range conns {
				dist[conn.addr.Addr] = int(conn.requests)
			}

			// 求熵
			entropy := mathx.CalcEntropy(dist)
			assert.True(t, entropy > test.threshold, fmt.Sprintf("熵：%f，小于：%f",
				entropy, test.threshold))
		})
	}
}

type mockClientConn struct{}

func (m mockClientConn) UpdateAddresses(addresses []resolver.Address) {
}

func (m mockClientConn) Connect() {
}
