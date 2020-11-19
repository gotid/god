package server_interceptors

import (
	"context"
	"git.zc0901.com/go/god/lib/load"
	"git.zc0901.com/go/god/lib/stat"
	"google.golang.org/grpc"
	"sync"
)

const serviceType = "rpc"

var (
	shedderStat *load.ShedderStat
	lock        sync.Mutex
)

// 一元卸流拦截器
func UnaryShedderInterceptor(shedder load.Shedder, metrics *stat.Metrics) grpc.UnaryServerInterceptor {
	ensureShedderStat()

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		shedderStat.IncrTotal()

		var promise load.Promise
		promise, err = shedder.Allow()
		if err != nil {
			metrics.AddDrop()
			shedderStat.IncrDrop()
			return
		}

		defer func() {
			if err == context.DeadlineExceeded {
				promise.Fail()
			} else {
				shedderStat.IncrPass()
				promise.Pass()
			}
		}()

		return handler(ctx, req)
	}
}

func ensureShedderStat() {
	lock.Lock()
	if shedderStat == nil {
		shedderStat = load.NewShedderStat(serviceType)
	}
	lock.Unlock()
}