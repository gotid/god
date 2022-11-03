package httpc

import (
	"context"
	"github.com/gotid/god/lib/breaker"
	"net/http"
)

type (
	// Service 表示一个远程 http 服务。
	Service interface {
		// Do 发送给定参数的 http 请求，并返回一个 http 响应。
		Do(ctx context.Context, method, url string, data interface{}) (*http.Response, error)
		// DoRequest 发送一个 http 请求，并返回一个 http 响应。
		DoRequest(r *http.Request) (*http.Response, error)
	}

	// Option 用于自定义 *http.Request。
	Option func(r *http.Request) *http.Request

	namedService struct {
		name string
		cli  *http.Client
		opts []Option
	}
)

// NewService 返回一个给定名称的远程服务 Service。
func NewService(name string, opts ...Option) Service {
	return NewServiceWithClient(name, http.DefaultClient, opts...)
}

// NewServiceWithClient 返回一个给定名称和客户端的远程服务 Service。
func NewServiceWithClient(name string, client *http.Client, opts ...Option) Service {
	return namedService{
		name: name,
		cli:  client,
		opts: opts,
	}
}

// Do 发送给定参数的 http 请求，并返回一个 http 响应。
func (s namedService) Do(ctx context.Context, method, url string, data interface{}) (*http.Response, error) {
	req, err := buildRequest(ctx, method, url, data)
	if err != nil {
		return nil, err
	}

	return s.DoRequest(req)
}

// DoRequest 发送一个 http 请求，并返回一个 http 响应。
func (s namedService) DoRequest(r *http.Request) (*http.Response, error) {
	return request(r, s)
}

func (s namedService) do(r *http.Request) (resp *http.Response, err error) {
	for _, opt := range s.opts {
		r = opt(r)
	}

	brk := breaker.Get(s.name)
	err = brk.DoWithAcceptable(func() error {
		resp, err = s.cli.Do(r)
		return err
	}, func(err error) bool {
		return err == nil && resp.StatusCode < http.StatusInternalServerError
	})

	return
}
