package httpx

import (
	"encoding/json"
	"github.com/gotid/god/api/internal/errcode"
	"github.com/gotid/god/api/internal/header"
	"github.com/gotid/god/lib/logx"
	"net/http"
	"sync"
)

var (
	errorHandler func(error) (int, interface{})
	lock         sync.RWMutex
)

// Error 将错误写入到响应编写器。
func Error(w http.ResponseWriter, err error, fns ...func(w http.ResponseWriter, err error)) {
	lock.RLock()
	handler := errorHandler
	lock.RUnlock()

	if handler == nil {
		if len(fns) > 0 {
			fns[0](w, err)
		} else if errcode.IsGrpcError(err) {
			// 不要对错误进行解包，也不要获取 status.Messages()，
			// 因为错误中包含了 rpc 的错误标头。
			http.Error(w, err.Error(), errcode.CodeFromGrpcError(err))
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		return
	}

	code, body := handler(err)
	if body == nil {
		w.WriteHeader(code)
		return
	}

	e, ok := body.(error)
	if ok {
		http.Error(w, e.Error(), code)
	} else {
		WriteJson(w, code, body)
	}
}

// Ok 将成功状态码写入响应编写器。
func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// OkJson 将 json 响应体及成功状态码写入响应编写器。
func OkJson(w http.ResponseWriter, body interface{}) {
	WriteJson(w, http.StatusOK, body)
}

// WriteJson 将响应体及状态码写入响应编写器。
func WriteJson(w http.ResponseWriter, code int, body interface{}) {
	bs, err := json.Marshal(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(ContentType, header.JsonContentType)
	w.WriteHeader(code)

	if n, err := w.Write(bs); err != nil {
		// http.ErrHandlerTimeout 已经被 http.TimeoutHandler 处理了，
		// 故此处忽略。
		if err != http.ErrHandlerTimeout {
			logx.Errorf("写入响应失败，错误：%s", err)
		}
	} else if n < len(bs) {
		logx.Errorf("实际字节数：%d，已写入字节数：%d", len(bs), n)
	}
}

// SetErrorHandler 设置错误处理器。
func SetErrorHandler(handler func(error) (int, interface{})) {
	lock.Lock()
	defer lock.Unlock()
	errorHandler = handler
}
