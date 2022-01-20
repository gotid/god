package handler

import (
	"net/http"
	"time"

	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/api/internal/security"
	"github.com/gotid/god/lib/codec"
	"github.com/gotid/god/lib/logx"
)

type UnsignedCallback func(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool, code int) // 未签名回调函数

// ContentSecurityHandler 返回一个鉴证内容安全的中间件。
func ContentSecurityHandler(decrypters map[string]codec.RsaDecryptor, tolerance time.Duration, strict bool, callbacks ...UnsignedCallback) func(handler http.Handler) http.Handler {
	if len(callbacks) == 0 {
		callbacks = append(callbacks, handleVerificationFailure)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodDelete, http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch:
				header, err := security.ParseContentSecurity(decrypters, r)
				if err != nil {
					logx.Errorf("签名解析失败，X-Content-Security: %s, 错误: %s",
						r.Header.Get(httpx.ContentSecurity), err.Error())
					executeCallbacks(w, r, next, strict, httpx.CodeSignatureInvalidHeader, callbacks)
				} else if code := security.VerifySignature(r, header, tolerance); code != httpx.CodeSignaturePass {
					logx.Errorf("签名校验失败, X-Content-Security: %s", r.Header.Get(httpx.ContentSecurity))
					executeCallbacks(w, r, next, strict, code, callbacks)
				} else if r.ContentLength > 0 && header.Encrypted() {
					CryptionHandler(header.Key)(next).ServeHTTP(w, r)
				} else {
					next.ServeHTTP(w, r)
				}
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

func executeCallbacks(w http.ResponseWriter, r *http.Request, next http.Handler,
	strict bool, code int, callbacks []UnsignedCallback) {
	for _, callback := range callbacks {
		callback(w, r, next, strict, code)
	}
}

func handleVerificationFailure(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool, code int) {
	if strict {
		w.WriteHeader(http.StatusForbidden)
	} else {
		next.ServeHTTP(w, r)
	}
}
