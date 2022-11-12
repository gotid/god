package handler

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gotid/god/api/internal/response"
	"github.com/gotid/god/api/token"
	"github.com/gotid/god/lib/logx"
	"net/http"
	"net/http/httputil"
)

const (
	noDetailReason = "无明确原因"
	jwtAudience    = "aud"
	jwtExpire      = "exp"
	jwtId          = "jti"
	jwtIssueAt     = "iat"
	jwtIssuer      = "iss"
	jwtNotBefore   = "nbf"
	jwtSubject     = "sub"
)

var (
	errInvalidToken = errors.New("身份校验令牌无效")
	errNoClaims     = errors.New("无鉴权参数")
)

type (
	// AuthorizeOptions 是一个授权选项。
	AuthorizeOptions struct {
		PrevSecret string
		Callback   UnauthorizedCallback
	}

	// UnauthorizedCallback 定义了未授权的回调方法。
	UnauthorizedCallback func(w http.ResponseWriter, r *http.Request, err error)

	// AuthorizeOption 自定义授权选项 AuthorizeOptions 的方法。
	AuthorizeOption func(opts *AuthorizeOptions)
)

// Authorize 返回一个 jwt 鉴权中间件。
func Authorize(secret string, opts ...AuthorizeOption) func(handler http.Handler) http.Handler {
	parser := token.NewParser()
	var authOpts AuthorizeOptions
	for _, opt := range opts {
		opt(&authOpts)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok, err := parser.ParseToken(r, secret, authOpts.PrevSecret)
			if err != nil {
				unauthorized(w, r, err, authOpts.Callback)
				return
			}

			if !tok.Valid {
				unauthorized(w, r, errInvalidToken, authOpts.Callback)
				return
			}

			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				unauthorized(w, r, errNoClaims, authOpts.Callback)
				return
			}

			ctx := r.Context()
			for k, v := range claims {
				switch k {
				case jwtAudience, jwtExpire, jwtId, jwtIssueAt, jwtIssuer, jwtNotBefore, jwtSubject:
				// 忽略标准声明
				default:
					ctx = context.WithValue(ctx, k, v)
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithPrevSecret 自定义上一次使用的秘钥
func WithPrevSecret(secret string) AuthorizeOption {
	return func(opts *AuthorizeOptions) {
		opts.PrevSecret = secret
	}
}

// WithUnauthorizedCallback 自定义鉴权失败回调函数。
func WithUnauthorizedCallback(callback UnauthorizedCallback) AuthorizeOption {
	return func(opts *AuthorizeOptions) {
		opts.Callback = callback
	}
}

func unauthorized(w http.ResponseWriter, r *http.Request, err error, callback UnauthorizedCallback) {
	writer := response.NewHeaderOnceResponseWriter(w)

	if err != nil {
		detailAuthLog(r, err.Error())
	} else {
		detailAuthLog(r, noDetailReason)
	}

	if callback != nil {
		callback(writer, r, err)
	}

	writer.WriteHeader(http.StatusUnauthorized)
}

func detailAuthLog(r *http.Request, reason string) {
	// 丢弃导出错误，仅用于调试
	details, _ := httputil.DumpRequest(r, true)
	logx.Errorf("授权失败：%s\n=> %+v", reason, string(details))
}
