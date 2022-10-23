package serverinterceptors

import (
	"context"
	"github.com/gotid/god/lib/load"
	"github.com/gotid/god/lib/stat"
	"google.golang.org/grpc"
	"sync"
)

const serviceType = "rpc"

var (
	lock         sync.Mutex
	sheddingStat *load.SheddingStat
)

// UnarySheddingInterceptor 用于一元请求的自动降载拦截器。
func UnarySheddingInterceptor(shedder load.Shedder, metrics *stat.Metrics) grpc.UnaryServerInterceptor {
	ensureSheddingStat()

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		sheddingStat.IncrTotal()
		var promise load.Promise
		promise, err = shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			sheddingStat.IncrDrop()
			return
		}

		defer func() {
			if err == context.DeadlineExceeded {
				promise.Fail()
			} else {
				sheddingStat.IncrPass()
				promise.Pass()
			}
		}()

		return handler(ctx, req)
	}
}

func ensureSheddingStat() {
	lock.Lock()
	if sheddingStat == nil {
		sheddingStat = load.NewSheddingStat(serviceType)
	}
	lock.Unlock()
}
