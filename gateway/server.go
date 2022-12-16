package gateway

import (
	"context"
	"fmt"
	"github.com/jhump/protoreflect/grpcreflect"
	"net/http"
	"strings"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gotid/god/api"
	"github.com/gotid/god/api/httpx"
	"github.com/gotid/god/gateway/internal"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mr"
	"github.com/gotid/god/rpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type (
	// Server 是一个基于 api.Server 的网关服务器。
	Server struct {
		*api.Server
		upstreams     []Upstream
		timeout       time.Duration
		processHeader func(header http.Header) []string
	}

	// Option 定义自定义 Server 的方法。
	Option func(server *Server)
)

// MustNewServer 返回一个新的网关服务器。
func MustNewServer(c Config, opts ...Option) *Server {
	svr := &Server{
		Server:    api.MustNewServer(c.Config),
		upstreams: c.Upstreams,
		timeout:   c.Timeout,
	}
	for _, opt := range opts {
		opt(svr)
	}

	return svr
}

// Start 启动网关服务器。
func (s *Server) Start() {
	logx.Must(s.build())
	s.Server.Start()
}

// Stop 停止网关服务器。
func (s *Server) Stop() {
	s.Server.Stop()
}

// 并行从上游 rpc 中添加服务器路由
func (s *Server) build() error {
	if err := s.ensureUpstreamNames(); err != nil {
		return err
	}

	return mr.MapReduceVoid(func(source chan<- any) {
		for _, upstream := range s.upstreams {
			source <- upstream
		}
	}, func(item any, writer mr.Writer, cancel func(error)) {
		upstream := item.(Upstream)
		client := rpc.MustNewClient(upstream.Grpc)
		source, err := s.createDescriptorSource(client, upstream)
		if err != nil {
			cancel(fmt.Errorf("%s：%w", upstream.Name, err))
			return
		}

		methods, err := internal.GetMethods(source)
		if err != nil {
			cancel(fmt.Errorf("%s：%w", upstream.Name, err))
			return
		}

		resolver := grpcurl.AnyResolverFromDescriptorSource(source)
		for _, m := range methods {
			if len(m.HttpMethod) > 0 && len(m.HttpPath) > 0 {
				writer.Write(api.Route{
					Method:  m.HttpMethod,
					Path:    m.HttpPath,
					Handler: s.buildHandler(source, resolver, client, m.RpcPath),
				})
			}
		}

		methodSet := make(map[string]struct{})
		for _, m := range methods {
			methodSet[m.RpcPath] = struct{}{}
		}
		for _, m := range upstream.Mappings {
			if _, ok := methodSet[m.RpcPath]; !ok {
				cancel(fmt.Errorf("%s：rpc 方法 %s 未找到", upstream.Name, m.RpcPath))
				return
			}

			writer.Write(api.Route{
				Method:  strings.ToUpper(m.Method),
				Path:    m.Path,
				Handler: s.buildHandler(source, resolver, client, m.RpcPath),
			})
		}
	}, func(pipe <-chan any, cancel func(error)) {
		for item := range pipe {
			route := item.(api.Route)
			s.Server.AddRoute(route)
		}
	})
}

func (s *Server) ensureUpstreamNames() error {
	for _, upstream := range s.upstreams {
		target, err := upstream.Grpc.BuildTarget()
		if err != nil {
			return err
		}

		upstream.Name = target
	}

	return nil
}

func (s *Server) createDescriptorSource(client rpc.Client, upstream Upstream) (grpcurl.DescriptorSource, error) {
	var source grpcurl.DescriptorSource
	var err error

	if len(upstream.ProtoSets) > 0 {
		source, err = grpcurl.DescriptorSourceFromProtoSets(upstream.ProtoSets...)
		if err != nil {
			return nil, err
		}
	} else {
		refCli := grpc_reflection_v1alpha.NewServerReflectionClient(client.Conn())
		client := grpcreflect.NewClient(context.Background(), refCli)
		source = grpcurl.DescriptorSourceFromServer(context.Background(), client)
	}

	return source, nil
}

func (s *Server) buildHandler(source grpcurl.DescriptorSource, resolver jsonpb.AnyResolver,
	client rpc.Client, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parser, err := internal.NewRequestParser(r, resolver)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		timeout := internal.GetTimeout(r.Header, s.timeout)
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		w.Header().Set(httpx.ContentType, httpx.JsonContentType)
		handler := internal.NewEventHandler(w, resolver)
		if err = grpcurl.InvokeRPC(ctx, source, client.Conn(), path, s.prepareMetadata(r.Header),
			handler, parser.Next); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		}

		status := handler.Status
		if status.Code() != codes.OK {
			httpx.ErrorCtx(r.Context(), w, status.Err())
		}
	}
}

func (s *Server) prepareMetadata(header http.Header) []string {
	headers := internal.ProcessHeaders(header)
	if s.processHeader != nil {
		headers = append(headers, s.processHeader(header)...)
	}

	return headers
}
