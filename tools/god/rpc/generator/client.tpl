{{.head}}

package {{.filePackage}}

import (
    "context"

    {{.pbPackage}}
    {{if ne .pbPackage .protoGoPackage}}{{.protoGoPackage}}{{end}}

    "github.com/gotid/god/rpc"
    "google.golang.org/grpc"
)

type (
    {{.alias}}

    {{.serviceName}} interface {
        {{.interface}}
    }

    default{{.serviceName}} struct {
        cli rpc.Client
    }
)

func New{{.serviceName}}(cli rpc.Client) {{.serviceName}} {
    return &default{{.serviceName}}{
        cli: cli,
    }
}

{{.functions}}