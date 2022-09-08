package errcode

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsGrpcError 判断指定 error 是否为一个 gRPC 错误。
func IsGrpcError(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(interface {
		GRPCStatus() *status.Status
	})

	return ok
}

// CodeFromGrpcError 将 gRPC 错误转为一个 HTTP 状态码。
// See: https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
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
