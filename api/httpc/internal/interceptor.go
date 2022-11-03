package internal

import "net/http"

type (
	Interceptor     func(*http.Request) (*http.Request, ResponseHandler)
	ResponseHandler func(*http.Response, error)
)
