package cors

import "net/http"

const (
	allowOrigin      = "Access-Control-Allow-Origin"
	allOrigins       = "*"
	allowMethods     = "Access-Control-Allow-Methods"
	allowHeaders     = "Access-Control-Allow-Headers"
	allowCredentials = "Access-Control-Allow-Credentials"
	exposeHeaders    = "Access-Control-Expose-Headers"
	requestMethod    = "Access-Control-Request-Method"
	requestHeaders   = "Access-Control-Request-Headers"
	allowHeadersVal  = "Content-Type, Origin, X-CSRF-Token, Authorization, AccessToken, Token, Range"
	exposeHeadersVal = "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers"
	methods          = "GET, HEAD, POST, PATCH, PUT, DELETE"
	allowTrue        = "true"
	maxAgeHeader     = "Access-Control-Max-Age"
	maxAgeHeaderVal  = "86400"
	varyHeader       = "Vary"
	originHeader     = "Origin"
)

// NotAllowedHandler 处理不允许的跨域请求。
// 仅允许一个 origin 源站或允许所有来源。
func NotAllowedHandler(origin ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		checkAndSetHeader(w, r, origin)

		if r.Method != http.MethodOptions {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

// Middleware 返回一个添加了 CORS 响应头的中间件。
func Middleware(origin ...string) func(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			checkAndSetHeader(w, r, origin)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
			} else {
				next(w, r)
			}
		}
	}
}

func checkAndSetHeader(w http.ResponseWriter, r *http.Request, origins []string) {
	setVaryHeaders(w, r)

	if len(origins) == 0 {
		setHeader(w, allOrigins)
		return
	}

	origin := r.Header.Get(originHeader)
	if isOriginAllowed(origins, origin) {
		setHeader(w, origin)
	}
}

func isOriginAllowed(allows []string, origin string) bool {
	for _, o := range allows {
		if o == allOrigins {
			return true
		}

		if o == origin {
			return true
		}

	}
	return false
}

func setHeader(w http.ResponseWriter, origin string) {
	header := w.Header()
	header.Set(allowOrigin, origin)
	header.Set(allowMethods, methods)
	header.Set(allowHeaders, allowHeadersVal)
	header.Set(exposeHeaders, exposeHeadersVal)
	if origin != allOrigins {
		header.Set(allowCredentials, allowTrue)
	}
	header.Set(maxAgeHeader, maxAgeHeaderVal)
}

func setVaryHeaders(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Add(varyHeader, originHeader)
	if r.Method == http.MethodOptions {
		header.Add(varyHeader, requestMethod)
		header.Add(varyHeader, requestHeaders)
	}
}