package router

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"git.zc0901.com/go/god/api/internal/context"
	"git.zc0901.com/go/god/lib/search"
)

const (
	allowHeader          = "Allow"
	allowMethodSeparator = ", "
)

var (
	ErrInvalidMethod = errors.New("不是有效的 http 请求方法")
	ErrInvalidPath   = errors.New("路径必须以 / 开头")
)

// Router 继承于 http.Handler 的路由器
type Router interface {
	http.Handler
	Handle(method, path string, handler http.Handler) error
	SetNotFoundHandler(handler http.Handler)
	SetNotAllowedHandler(handler http.Handler)
}

// 路由处理器
type router struct {
	trees      map[string]*search.Tree
	notFound   http.Handler
	notAllowed http.Handler
}

func NewRouter() Router {
	return &router{trees: make(map[string]*search.Tree)}
}

func (rt *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqPath := path.Clean(r.URL.Path)
	if tree, ok := rt.trees[r.Method]; ok {
		if result, ok := tree.Search(reqPath); ok {
			if len(result.Params) > 0 {
				r = context.WithPathVars(r, result.Params)
			}
			result.Item.(http.Handler).ServeHTTP(w, r)
			return
		}
	}

	allows, ok := rt.methodsNotAllow(r.Method, reqPath)
	if !ok {
		rt.handleNotFound(w, r)
		return
	}

	if rt.notAllowed != nil {
		rt.notAllowed.ServeHTTP(w, r)
	} else {
		w.Header().Set(allowHeader, allows)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (rt *router) Handle(method, reqPath string, handler http.Handler) error {
	if !validMethod(method) {
		return ErrInvalidMethod
	}

	if len(reqPath) == 0 || reqPath[0] != '/' {
		return ErrInvalidPath
	}

	cleanPath := path.Clean(reqPath)
	if tree, ok := rt.trees[method]; ok {
		return tree.Add(cleanPath, handler)
	} else {
		tree := search.NewTree()
		rt.trees[method] = tree
		return tree.Add(cleanPath, handler)
	}
}

// 设置资源未找到处理器
func (rt *router) SetNotFoundHandler(handler http.Handler) {
	rt.notFound = handler
}

// 设置资源不允许访问处理器
func (rt *router) SetNotAllowedHandler(handler http.Handler) {
	rt.notAllowed = handler
}

// 检测指定的请求路径是否可使用指定的请求方法
func (rt *router) methodsNotAllow(method, reqPath string) (string, bool) {
	var allows []string

	for treeMethod, tree := range rt.trees {
		if treeMethod == method {
			continue
		}

		_, ok := tree.Search(reqPath)
		if ok {
			allows = append(allows, treeMethod)
		}
	}

	if len(allows) > 0 {
		return strings.Join(allows, allowMethodSeparator), true
	} else {
		return "", false
	}
}

// 处理请求方法-路径未找到的情况
func (rt *router) handleNotFound(w http.ResponseWriter, r *http.Request) {
	if rt.notFound != nil {
		rt.notFound.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func validMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodPost ||
		method == http.MethodPut || method == http.MethodOptions ||
		method == http.MethodDelete || method == http.MethodPatch ||
		method == http.MethodHead
}
