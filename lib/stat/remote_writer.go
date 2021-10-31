package stat

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"git.zc0901.com/go/god/lib/logx"
)

const httpTimeout = 5 * time.Second

var ErrWriteFailed = errors.New("远程统计提交错误")

type RemoteWriter struct {
	endpoint string
}

// NewRemoteWriter 新建远程上报器。
func NewRemoteWriter(endpoint string) Writer {
	return &RemoteWriter{endpoint}
}

func (w RemoteWriter) Write(report *ReportItem) error {
	bs, err := json.Marshal(report)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: httpTimeout}
	// endpoint 就是推送到 prometheus 或 opentelemetry 服务器的地址
	resp, err := client.Post(w.endpoint, "application/json", bytes.NewReader(bs))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logx.Errorf("提交统计报告失败，错误码：%d, 原因：%s", resp.StatusCode, resp.Status)
		return ErrWriteFailed
	}

	return nil
}
