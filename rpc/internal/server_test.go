package internal

import (
	"github.com/gotid/god/lib/stat"
	"github.com/gotid/god/rpc/internal/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"sync"
	"testing"
)

func TestServer(t *testing.T) {
	metrics := stat.NewMetrics("foo")
	s := NewServer("localhost:54321", WithMetrics(metrics))
	s.SetName("mock")
	var wg sync.WaitGroup
	var grpcServer *grpc.Server
	var lock sync.Mutex

	wg.Add(1)
	go func() {
		err := s.Start(func(server *grpc.Server) {
			lock.Lock()
			mock.RegisterDepositServiceServer(server, new(mock.DepositServer))
			grpcServer = server
			lock.Unlock()
			wg.Done()
		})
		assert.Nil(t, err)
	}()
	wg.Wait()

	lock.Lock()
	grpcServer.GracefulStop()
	lock.Unlock()
}

func TestServer_WithBadAddress(t *testing.T) {
	s := NewServer("localhost:111111")
	s.SetName("mock")
	err := s.Start(func(server *grpc.Server) {
		mock.RegisterDepositServiceServer(server, new(mock.DepositServer))
	})
	assert.NotNil(t, err)
}

func TestWithStreamServerInterceptors(t *testing.T) {
	opts := WithStreamServerInterceptors()
	assert.NotNil(t, opts)
}

func TestWithUnaryServerInterceptors(t *testing.T) {
	opts := WithUnaryServerInterceptors()
	assert.NotNil(t, opts)
}
