package chain

import "net/http"

type (
	// Chain 接口定义了一个中间件链。
	Chain interface {
		// Append 将给定中间件列表附加到已有中间件的后面。
		Append(middlewares ...Middleware) Chain
		// Prepend 将给定中间件预置到已有中间件的前面。
		Prepend(middlewares ...Middleware) Chain
		// Then 将中间件按顺序应用到处理器 h 上，并返回最终的处理器。
		Then(h http.Handler) http.Handler
		ThenFunc(fn http.HandlerFunc) http.Handler
	}

	// Middleware 是一个 http 中间件。
	Middleware func(http.Handler) http.Handler

	// chain 作为一个 http.Handler 中间件列表。
	// chain 实际上是不变的：
	// 一旦创建，它将始终保持以相同的顺序保存同一组中间件。
	chain struct {
		middlewares []Middleware
	}
)

// New 创建一个新的 Chain，记住给定的中间件顺序。
// New 不提供其他功能，中间件只能通过 Then() 或 ThenFunc() 进行调用。
func New(middlewares ...Middleware) Chain {
	return chain{middlewares: append(([]Middleware)(nil), middlewares...)}
}

func (c chain) Append(middlewares ...Middleware) Chain {
	return chain{middlewares: join(c.middlewares, middlewares)}
}

func (c chain) Prepend(middlewares ...Middleware) Chain {
	return chain{middlewares: join(middlewares, c.middlewares)}
}

func (c chain) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}

	for i := range c.middlewares {
		h = c.middlewares[len(c.middlewares)-1-i](h)
	}

	return h
}

// ThenFunc 将中间件按顺序应用到处理函数 fn 上，并返回中的处理器。
func (c chain) ThenFunc(fn http.HandlerFunc) http.Handler {
	// 该 nil 检测不能去掉，因为 Go 中存在 "nil is not nil" 的问题。
	// Required due to: https://stackoverflow.com/questions/33426977/how-to-golang-check-a-variable-is-nil
	if fn == nil {
		return c.Then(nil)
	}

	return c.Then(fn)
}

func join(a, b []Middleware) []Middleware {
	ms := make([]Middleware, 0, len(a)+len(b))
	ms = append(ms, a...)
	ms = append(ms, b...)
	return ms
}
