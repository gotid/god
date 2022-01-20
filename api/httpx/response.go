package httpx

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gotid/god/lib/logx"
)

var (
	errorHandler  func(error) (int, interface{})
	okJsonHandler func(body interface{}) interface{}
	lock          sync.RWMutex
)

type Message struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"message"`
}

func (e *Message) Error() string {
	return e.Msg
}

func NewCodeError(code int, msg string) error {
	return &Message{Code: code, Msg: msg}
}

func NewDefaultError(msg string) error {
	return NewCodeError(0, msg)
}

// Error 错误响应，支持自定义错误处理器
func Error(w http.ResponseWriter, err error) {
	lock.RLock()
	handler := errorHandler
	lock.RUnlock()

	if handler == nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code, body := handler(err)
	if body == nil {
		w.WriteHeader(code)
		return
	}

	e, ok := body.(error)
	if ok {
		WriteJson(w, http.StatusOK, &Message{
			Code: code,
			Msg:  e.Error(),
		})
	} else {
		WriteJson(w, code, body)
	}
}

// Ok 正常响应
func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// OkJson 正常JSON响应
func OkJson(w http.ResponseWriter, body interface{}) {
	lock.RLock()
	handler := okJsonHandler
	lock.RUnlock()

	if handler != nil {
		body = okJsonHandler(body)
	}

	WriteJson(w, http.StatusOK, body)
}

// SetErrorHandler 设置自定义错误处理器
func SetErrorHandler(handler func(error) (int, interface{})) {
	lock.Lock()
	defer lock.Unlock()
	errorHandler = handler
}

// SetOkJsonHandler 设置自定义成功处理器
func SetOkJsonHandler(handler func(body interface{}) interface{}) {
	lock.Lock()
	defer lock.Unlock()
	okJsonHandler = handler
}

// WriteJson 写JSON响应
func WriteJson(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set(ContentType, ApplicationJson)
	w.WriteHeader(code)

	if bs, err := json.Marshal(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if n, err := w.Write(bs); err != nil {
		// http.ErrHandlerTimeout 已经被 http.TimeoutHandler 处理了
		// 所以此处忽略。
		if err != http.ErrHandlerTimeout {
			logx.Errorf("写响应失败，错误：%s", err)
		}
	} else if n < len(bs) {
		logx.Errorf("实际字节数：%d，写字节数：%d", len(bs), n)
	}
}
