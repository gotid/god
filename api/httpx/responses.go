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
	errorHandler  func(error) (int, any)
	okJsonHandler func(body any) any
	lock          sync.RWMutex
)

// Message 响应消息
type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func (e *Message) Error() string {
	return e.Message
}

// NewCodeError 返回一个指定代码和消息的错误。
func NewCodeError(code int, msg string) error {
	return &Message{Code: code, Message: msg}
}

// NewDefaultError 返回一个代码为0的默认错误。
func NewDefaultError(msg string) error {
	return NewCodeError(0, msg)
}

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
func OkJson(w http.ResponseWriter, body any) {
	lock.RLock()
	handler := okJsonHandler
	lock.RUnlock()

	if handler != nil {
		body = handler(body)
	}

	WriteJson(w, http.StatusOK, body)
}

// WriteJson 将响应体及状态码写入响应编写器。
func WriteJson(w http.ResponseWriter, code int, body any) {
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
func SetErrorHandler(handler func(error) (int, any)) {
	lock.Lock()
	defer lock.Unlock()
	errorHandler = handler
}

// SetOkJsonHandler 设置自定义成功处理器
func SetOkJsonHandler(handler func(body any) any) {
	lock.Lock()
	defer lock.Unlock()
	okJsonHandler = handler
}
