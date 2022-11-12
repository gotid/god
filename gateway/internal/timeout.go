package internal

import (
	"net/http"
	"time"
)

const grpcTimeoutHeader = "Grpc-Timeout"

// GetTimeout 返回标头中的超时时长，如未设置则返回默认超时时长。
func GetTimeout(header http.Header, defaultTimeout time.Duration) time.Duration {
	if timeout := header.Get(grpcTimeoutHeader); len(timeout) > 0 {
		if t, err := time.ParseDuration(timeout); err == nil {
			return t
		}
	}

	return defaultTimeout
}
