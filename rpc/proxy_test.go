package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/gotid/god/rpc/internal/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	mock.RegisterDepositServiceServer(server, &mock.DepositServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestProxy(t *testing.T) {
	tests := []struct {
		name    string
		amount  float32
		res     *mock.DepositResponse
		errCode codes.Code
		errMsg  string
	}{
		{
			name:    "带负数的无效请求",
			amount:  -1.11,
			res:     nil,
			errCode: codes.InvalidArgument,
			errMsg:  fmt.Sprintf("不能存储 %v", -1.11),
		},
		{
			name:    "非负数的有效请求",
			amount:  0.00,
			res:     &mock.DepositResponse{Ok: true},
			errCode: codes.OK,
			errMsg:  "",
		},
	}

	proxy := NewProxy("foo", WithDialOption(grpc.WithInsecure()),
		WithDialOption(grpc.WithContextDialer(dialer())))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conn, err := proxy.TakeConn(context.Background())
			assert.Nil(t, err)
			cli := mock.NewDepositServiceClient(conn)
			request := &mock.DepositRequest{Amount: test.amount}
			response, err := cli.Deposit(context.Background(), request)
			if response != nil {
				assert.True(t, len(response.String()) > 0)
				if response.GetOk() != test.res.GetOk() {
					t.Error("响应：期待", test.res.GetOk(), "收到", response.GetOk())
				}
			}
			if err != nil {
				if e, ok := status.FromError(err); ok {
					if e.Code() != test.errCode {
						t.Error("错误码：期待", codes.InvalidArgument, "收到", e.Code())
					}
					if e.Message() != test.errMsg {
						t.Error("错误消息：期待", test.errMsg, "收到", e.Message())
					}
				}
			}
		})
	}
}
