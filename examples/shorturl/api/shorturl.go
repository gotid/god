package main

import (
	"flag"
	"fmt"
	"github.com/gotid/god/lib/store/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"

	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/lib/conf"

	"github.com/gotid/god/examples/shorturl/api/internal/config"
	"github.com/gotid/god/examples/shorturl/api/internal/handler"
	"github.com/gotid/god/examples/shorturl/api/internal/svc"
)

var configFile = flag.String("f", "etc/shorturl-api.yaml", "配置文件")

// CodeFromGrpcError 将 gRPC 错误转为一个 HTTP 状态码。
// 详见：https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func CodeFromGrpcError(err error) int {
	code := status.Code(err)
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Internal, codes.DataLoss, codes.Unknown:
		return http.StatusInternalServerError
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	}

	return http.StatusInternalServerError
}

// IsGrpcError 检查错误是否为一个 gRPC 的错误。
func IsGrpcError(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(interface {
		GRPCStatus() *status.Status
	})

	return ok
}

func main() {
	flag.Parse()

	// 设置错误处理函数
	httpx.SetErrorHandler(func(err error) (int, any) {
		if IsGrpcError(err) {
			fmt.Println("GRPC错误: ", err)
			statusCode := CodeFromGrpcError(err)
			msg := status.Convert(err).Message()
			if msg == sqlx.ErrNotFound.Error() {
				msg = "查无此项"
			}
			return statusCode, httpx.Message{
				Code:    -1,
				Message: msg,
			}
		}
		return http.StatusConflict, httpx.Message{
			Code:    -1,
			Message: err.Error(),
		}
	})
	httpx.SetOkJsonHandler(func(body any) any {
		return httpx.Message{
			Data: body,
		}
	})

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := api.MustNewServer(c.Config)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("启动 api 服务器 %s:%d...\n", c.Host, c.Port)
	server.Start()
}
