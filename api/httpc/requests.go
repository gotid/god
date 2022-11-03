package httpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gotid/god/api/httpc/internal"
	"github.com/gotid/god/api/internal/header"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/mapping"
	"github.com/gotid/god/lib/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"net/http/httptrace"
	nurl "net/url"
	"strings"
)

var interceptors = []internal.Interceptor{
	internal.LogInterceptor,
}

type (
	client interface {
		do(r *http.Request) (*http.Response, error)
	}

	defaultClient struct{}
)

func (c defaultClient) do(r *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(r)
}

// Do 发送给定参数的 http 请求，并返回一个 http 响应。
// data 自动编组至一个 *httpRequest，通常是定义在一个 API 文件中。
func Do(ctx context.Context, method, url string, data interface{}) (*http.Response, error) {
	req, err := buildRequest(ctx, method, url, data)
	if err != nil {
		return nil, err
	}

	return DoRequest(req)
}

// DoRequest 发送一个 http 请求，并返回一个 http 响应。
func DoRequest(r *http.Request) (*http.Response, error) {
	return request(r, defaultClient{})
}

func request(r *http.Request, cli client) (*http.Response, error) {
	tracer := otel.GetTracerProvider().Tracer(trace.Name)
	propagator := otel.GetTextMapPropagator()

	spanName := r.URL.Path
	ctx, span := tracer.Start(
		r.Context(),
		spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(semconv.HTTPClientAttributesFromHTTPRequest(r)...),
	)
	defer span.End()

	respHandlers := make([]internal.ResponseHandler, len(interceptors))
	for i, interceptor := range interceptors {
		var h internal.ResponseHandler
		r, h = interceptor(r)
		respHandlers[i] = h
	}

	clientTrace := httptrace.ContextClientTrace(ctx)
	if clientTrace != nil {
		ctx = httptrace.WithClientTrace(ctx, clientTrace)
	}

	r = r.WithContext(ctx)
	propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))

	resp, err := cli.do(r)
	for i := len(respHandlers) - 1; i >= 0; i-- {
		respHandlers[i](resp, err)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)
	span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(resp.StatusCode))

	return resp, err
}

// 构建请求体。
// 其中，data 为 json 结构的请求参数。
func buildRequest(ctx context.Context, method, url string, data interface{}) (*http.Request, error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	var m map[string]map[string]interface{}
	if data != nil {
		m, err = mapping.Marshal(data)
		if err != nil {
			return nil, err
		}
	}

	// 填充路径
	if err = fillPath(u, m[pathKey]); err != nil {
		return nil, err
	}

	// 读取 json 请求体
	var reader io.Reader
	jsonVars, hasJsonBody := m[jsonKey]
	if hasJsonBody {
		if method == http.MethodGet {
			return nil, ErrGetWithBody
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		if err = enc.Encode(jsonVars); err != nil {
			return nil, err
		}

		reader = &buf
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}

	// 构建编码后的表单查询字符串
	req.URL.RawQuery = buildFormQuery(u, m[formKey])

	// 填充请求头
	fillHeader(req, m[headerKey])
	if hasJsonBody {
		req.Header.Set(header.ContentType, header.JsonContentType)
	}

	return req, nil
}

func fillHeader(r *http.Request, m map[string]interface{}) {
	for k, v := range m {
		r.Header.Add(k, fmt.Sprint(v))
	}
}

func buildFormQuery(u *nurl.URL, m map[string]interface{}) string {
	query := u.Query()
	for k, v := range m {
		query.Add(k, fmt.Sprint(v))
	}

	return query.Encode()
}

func fillPath(u *nurl.URL, m map[string]interface{}) error {
	used := make(map[string]lang.PlaceholderType)
	fields := strings.Split(u.Path, slash)

	for i := range fields {
		field := fields[i]
		if len(field) > 0 && field[0] == colon {
			name := field[1:]
			v, ok := m[name]
			if !ok {
				return fmt.Errorf("缺少路径变量 %q", name)
			}
			value := fmt.Sprint(v)
			if len(value) == 0 {
				return fmt.Errorf("路径变量的值不能为空 %q", name)
			}
			fields[i] = value
			used[name] = lang.Placeholder
		}
	}

	if len(m) != len(used) {
		for key := range used {
			delete(m, key)
		}

		var unused []string
		for key := range m {
			unused = append(unused, key)
		}

		return fmt.Errorf("提供了用不到的路径变量：%q", strings.Join(unused, ", "))
	}

	u.Path = strings.Join(fields, slash)
	return nil
}
