package generator

import (
	conf "github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/rpc/parser"
	"github.com/gotid/god/tools/god/util/ctx"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/util/stringx"
	"path/filepath"
	"strings"
)

const (
	call     = "call"
	wd       = "wd"
	etc      = "etc"
	internal = "internal"
	config   = "config"
	logic    = "logic"
	server   = "server"
	svc      = "svc"
	pb       = "pb"
	protoGo  = "proto-go"
)

type (
	// Dir 定义一个目录。
	Dir struct {
		Base            string
		Filename        string
		Package         string
		GetChildPackage func(childPath string) (string, error)
	}

	// DirContext 接口定义一个 rpc 服务目录的上下文。
	DirContext interface {
		GetCall() Dir
		GetEtc() Dir
		GetInternal() Dir
		GetConfig() Dir
		GetLogic() Dir
		GetServer() Dir
		GetSvc() Dir
		GetPb() Dir
		GetProtoGo() Dir
		GetMain() Dir
		GetServiceName() stringx.String
		SetPbDir(pbDir, grpcDir string)
	}

	defaultDirContext struct {
		inner       map[string]Dir
		serviceName stringx.String
		ctx         *ctx.ProjectContext
	}
)

// Valid 判断目录是否有效。
func (d *Dir) Valid() bool {
	return len(d.Filename) > 0 && len(d.Package) > 0
}

func (d *defaultDirContext) GetCall() Dir {
	return d.inner[call]
}

func (d *defaultDirContext) GetEtc() Dir {
	return d.inner[etc]
}

func (d *defaultDirContext) GetInternal() Dir {
	return d.inner[internal]
}

func (d *defaultDirContext) GetConfig() Dir {
	return d.inner[config]
}

func (d *defaultDirContext) GetLogic() Dir {
	return d.inner[logic]
}

func (d *defaultDirContext) GetServer() Dir {
	return d.inner[server]
}

func (d *defaultDirContext) GetSvc() Dir {
	return d.inner[svc]
}

func (d *defaultDirContext) GetPb() Dir {
	return d.inner[pb]
}

func (d *defaultDirContext) GetProtoGo() Dir {
	return d.inner[protoGo]
}

// GetMain 获取工作空间 workDir。
func (d *defaultDirContext) GetMain() Dir {
	return d.inner[wd]
}

func (d *defaultDirContext) GetServiceName() stringx.String {
	return d.serviceName
}

func (d *defaultDirContext) SetPbDir(pbDir, grpcDir string) {
	d.inner[pb] = Dir{
		Filename: pbDir,
		Package:  filepath.ToSlash(filepath.Join(d.ctx.Path, strings.TrimPrefix(pbDir, d.ctx.Dir))),
		Base:     filepath.Base(pbDir),
	}
	d.inner[protoGo] = Dir{
		Filename: grpcDir,
		Package:  filepath.ToSlash(filepath.Join(d.ctx.Path, strings.TrimPrefix(grpcDir, d.ctx.Dir))),
		Base:     filepath.Base(grpcDir),
	}
}

func mkdir(ctx *ctx.ProjectContext, proto parser.Proto, _ *conf.Config, rpcCtx *RpcContext) (DirContext, error) {
	inner := make(map[string]Dir)
	etcDir := filepath.Join(ctx.WorkDir, "etc")
	clientDir := filepath.Join(ctx.WorkDir, "client")
	internalDir := filepath.Join(ctx.WorkDir, "internal")
	pbDir := filepath.Join(ctx.WorkDir, proto.GoPackage)
	configDir := filepath.Join(internalDir, "config")
	logicDir := filepath.Join(internalDir, "logic")
	serverDir := filepath.Join(internalDir, "server")
	svcDir := filepath.Join(internalDir, "svc")
	protoGoDir := pbDir
	if rpcCtx != nil {
		pbDir = rpcCtx.ProtoGenGrpcDir
		protoGoDir = rpcCtx.ProtoGenGoDir
	}

	getChildPackage := func(parent, childPath string) (string, error) {
		child := strings.TrimPrefix(childPath, parent)
		abs := filepath.Join(parent, strings.ToLower(child))
		if rpcCtx.Multiple {
			if err := pathx.MkdirIfNotExist(abs); err != nil {
				return "", err
			}
		}
		childPath = strings.TrimPrefix(abs, ctx.Dir)
		pkg := filepath.Join(ctx.Path, childPath)
		return filepath.ToSlash(pkg), nil
	}

	if !rpcCtx.Multiple {
		svcName := proto.Service[0].Name
		callDir := filepath.Join(ctx.WorkDir, strings.ToLower(stringx.From(svcName).ToCamel()))
		if strings.EqualFold(svcName, filepath.Base(proto.GoPackage)) {
			callDir = filepath.Join(ctx.WorkDir, strings.ToLower(stringx.From(svcName+"_client").ToCamel()))
		}
		inner[call] = Dir{
			Filename: callDir,
			Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(callDir, ctx.Dir))),
			Base:     filepath.Base(callDir),
			GetChildPackage: func(childPath string) (string, error) {
				return getChildPackage(callDir, childPath)
			},
		}
	} else {
		inner[call] = Dir{
			Filename: clientDir,
			Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(clientDir, ctx.Dir))),
			Base:     filepath.Base(clientDir),
			GetChildPackage: func(childPath string) (string, error) {
				return getChildPackage(clientDir, childPath)
			},
		}
	}

	inner[wd] = Dir{
		Filename: ctx.WorkDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(ctx.WorkDir, ctx.Dir))),
		Base:     filepath.Base(ctx.WorkDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(ctx.WorkDir, childPath)
		},
	}

	inner[etc] = Dir{
		Filename: etcDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(etcDir, ctx.Dir))),
		Base:     filepath.Base(etcDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(etcDir, childPath)
		},
	}
	inner[internal] = Dir{
		Filename: internalDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(internalDir, ctx.Dir))),
		Base:     filepath.Base(internalDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(internalDir, childPath)
		},
	}
	inner[config] = Dir{
		Filename: configDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(configDir, ctx.Dir))),
		Base:     filepath.Base(configDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(configDir, childPath)
		},
	}
	inner[logic] = Dir{
		Filename: logicDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(logicDir, ctx.Dir))),
		Base:     filepath.Base(logicDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(logicDir, childPath)
		},
	}
	inner[server] = Dir{
		Filename: serverDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(serverDir, ctx.Dir))),
		Base:     filepath.Base(serverDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(serverDir, childPath)
		},
	}
	inner[svc] = Dir{
		Filename: svcDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(svcDir, ctx.Dir))),
		Base:     filepath.Base(svcDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(svcDir, childPath)
		},
	}
	inner[pb] = Dir{
		Filename: pbDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(pbDir, ctx.Dir))),
		Base:     filepath.Base(pbDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(pbDir, childPath)
		},
	}
	inner[protoGo] = Dir{
		Filename: protoGoDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(protoGoDir, ctx.Dir))),
		Base:     filepath.Base(protoGoDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(protoGoDir, childPath)
		},
	}

	for _, v := range inner {
		err := pathx.MkdirIfNotExist(v.Filename)
		if err != nil {
			return nil, err
		}
	}

	serviceName := strings.TrimSuffix(proto.Name, filepath.Ext(proto.Name))
	return &defaultDirContext{
		ctx:         ctx,
		inner:       inner,
		serviceName: stringx.From(strings.ReplaceAll(serviceName, "-", "")),
	}, nil
}
