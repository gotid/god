package main

import (
    "flag"
    "fmt"

    {{.imports}}

    "github.com/gotid/god/lib/conf"
    "github.com/gotid/god/lib/service"
    "github.com/gotid/god/rpc"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/{{.serviceName}}.yaml", "配置文件")

func main() {
    flag.Parse()

    var c config.Config
    conf.MustLoad(*configFile, &c)
    ctx := svc.NewServiceContext(c)

    s := rpc.MustNewServer(c.ServerConfig, func(grpcServer *grpc.Server) {
{{range .serviceNames}}       {{.Pkg}}.Register{{.Service}}Server(grpcServer, {{.ServerPkg}}.New{{.Service}}Server(ctx))
{{end}}
        if c.Mode == service.DevMode || c.Mode == service.TestMode {
            reflection.Register(grpcServer)
        }
    })
    defer s.Stop()

    fmt.Printf("启动 rpc 服务器 %s...\n", c.ListenOn)
    s.Start()
}