package stat

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gotid/god/lib/logx"
	"net/http"
	"time"
)

const (
	httpTimeout     = 5 * time.Second
	jsonContentType = "application/json; charset=utf-8"
)

// ErrWriteFailed 是一个代表提交统计报告失败的错误。
var ErrWriteFailed = errors.New("提交失败")

// RemoteWriter 是一个编写 StatReport 的远程编写器。
type RemoteWriter struct {
	endpoint string
}

// NewRemoteWriter 返回一个 RemoteWriter。
func NewRemoteWriter(endpoint string) Writer {
	return &RemoteWriter{
		endpoint: endpoint,
	}
}

func (rw *RemoteWriter) Write(report *StatReport) error {
	bs, err := json.Marshal(report)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: httpTimeout,
	}
	resp, err := client.Post(rw.endpoint, jsonContentType, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logx.Errorf("编写报告失败，状态码：%d，原因：%s，网址：%s", resp.StatusCode, resp.Status, rw.endpoint)
		return ErrWriteFailed
	}

	return nil
}
