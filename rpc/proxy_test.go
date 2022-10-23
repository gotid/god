package rpc

import (
	"context"
	"fmt"
	"github.com/gotid/god/rpc/internal/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"testing"
)

func TestProxy(t *testing.T) {
	tests := []struct {
		name    string
		amount  float32
		res     *mock.DepositResponse
		errCode codes.Code
		errMsg  string
	}{
		{
			"金额为负数的无效请求",
			-1.11,
			nil,
			codes.InvalidArgument,
			fmt.Sprintf("无法存款 %v", -1.11),
		},
		{
			"金额为非负数的有效请求",
			0.00,
			&mock.DepositResponse{Ok: true},
			codes.OK,
			"",
		},
	}

	proxy := NewProxy("foo", WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		WithDialOption(grpc.WithContextDialer(dialer())))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := proxy.TakeConn(context.Background())
			assert.Nil(t, err)
			cli := mock.NewDepositServiceClient(conn)
			request := &mock.DepositRequest{Amount: tt.amount}
			response, err := cli.Deposit(context.Background(), request)
			if response != nil {
				assert.True(t, len(response.String()) > 0)
				if response.GetOk() != tt.res.GetOk() {
					t.Error("response: expected", tt.res.GetOk(), "received", response.GetOk())
				}
			}
			if err != nil {
				if e, ok := status.FromError(err); ok {
					if e.Code() != tt.errCode {
						t.Error("error code: expected", codes.InvalidArgument, "received", e.Code())
					}
					if e.Message() != tt.errMsg {
						t.Error("错误消息：期望", tt.errMsg, "收到", e.Message())
					}
				}
			}
		})
	}
}

func TestProxy_TakeConnNewClientFailed(t *testing.T) {
	proxy := NewProxy("foo", WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		WithDialOption(grpc.WithBlock()))
	_, err := proxy.TakeConn(context.Background())
	assert.NotNil(t, err)
}
