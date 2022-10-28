package rpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gotid/god/lib/discov"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/rpc/internal/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func init() {
	logx.Disable()
}

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

func TestDepositServer_Deposit(t *testing.T) {
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
		{
			"valid request with long handling time",
			2000.00,
			nil,
			codes.DeadlineExceeded,
			"context deadline exceeded",
		},
	}

	directClient := MustNewClient(
		ClientConfig{
			Endpoints: []string{"foo"},
			App:       "foo",
			Token:     "bar",
			Timeout:   1000,
		},
		WithDialOption(grpc.WithContextDialer(dialer())),
		WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	nonBlockClient := MustNewClient(
		ClientConfig{
			Endpoints: []string{"foo"},
			App:       "foo",
			Token:     "bar",
			Timeout:   1000,
			NonBlock:  true,
		},
		WithDialOption(grpc.WithContextDialer(dialer())),
		WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	tarConfClient := MustNewClient(
		ClientConfig{
			Target:  "foo",
			App:     "foo",
			Token:   "bar",
			Timeout: 1000,
		},
		WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		WithDialOption(grpc.WithContextDialer(dialer())),
		WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	targetClient, err := NewClientWithTarget("foo",
		WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		WithDialOption(grpc.WithContextDialer(dialer())), WithUnaryClientInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
				invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				return invoker(ctx, method, req, reply, cc, opts...)
			}), WithTimeout(1000*time.Millisecond))
	assert.Nil(t, err)
	clients := []Client{
		directClient,
		nonBlockClient,
		tarConfClient,
		targetClient,
	}
	SetClientSlowThreshold(time.Second)

	for _, tt := range tests {
		tt := tt
		for _, client := range clients {
			client := client
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				cli := mock.NewDepositServiceClient(client.Conn())
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
							t.Error("error message: expected", tt.errMsg, "received", e.Message())
						}
					}
				}
			})
		}
	}
}

func TestNewClientWithError(t *testing.T) {
	_, err := NewClient(
		ClientConfig{
			App:     "foo",
			Token:   "bar",
			Timeout: 1000,
		},
		WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		WithDialOption(grpc.WithContextDialer(dialer())),
		WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	assert.NotNil(t, err)
}

func TestEtcdError(t *testing.T) {
	_, err := NewClient(
		ClientConfig{
			Etcd: discov.EtcdConfig{
				Hosts: []string{"localhost:2379"},
				Key:   "mock",
			},
			App:     "foo",
			Token:   "bar",
			Timeout: 1,
		},
		WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		WithDialOption(grpc.WithContextDialer(dialer())),
		WithUnaryClientInterceptor(func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	assert.NotNil(t, err)
}
