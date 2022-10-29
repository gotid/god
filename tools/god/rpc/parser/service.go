package parser

import (
	"errors"
	"fmt"
	"github.com/emicklei/proto"
	"path/filepath"
	"strings"
)

type (
	// Service 表示 rpc 服务，是解析 proto 文件后得到的相应内容。
	Service struct {
		*proto.Service
		RPC []*RPC
	}

	// Services 是一个 Service 切片。
	Services []Service
)

func (s Services) validate(filename string, multipleOpt ...bool) error {
	if len(s) == 0 {
		return errors.New("未找到 rpc 服务")
	}

	var multiple bool
	for _, c := range multipleOpt {
		multiple = c
	}

	if !multiple && len(s) > 1 {
		return errors.New("单服务模式下，proto 文件只能有一个 service。\n多服务请使用参数 -m")
	}

	name := filepath.Base(filename)
	for _, service := range s {
		for _, rpc := range service.RPC {
			if strings.Contains(rpc.RequestType, ".") {
				return fmt.Errorf("行 %v:%v，请求类型必须定义在 %s",
					rpc.Position.Line, rpc.Position.Column, name)
			}
			if strings.Contains(rpc.ReturnsType, ".") {
				return fmt.Errorf("行 %v:%v，返回类型必须定义在 %s",
					rpc.Position.Line, rpc.Position.Column, name)
			}
		}
	}

	return nil
}
