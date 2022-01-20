package load

import (
	"io"

	"github.com/gotid/god/lib/syncx"
)

type
// ShedderGroup 是一个基于键名的泄流阀管理器。
ShedderGroup struct {
	options []ShedderOption
	manager *syncx.ResourceManager
}

// NewShedderGroup 返回一个泄流阀管理器。
func NewShedderGroup(opts ...ShedderOption) *ShedderGroup {
	return &ShedderGroup{
		options: opts,
		manager: syncx.NewResourceManager(),
	}
}

// GetShedder 获取指定键名的可复用泄流阀。
func (g *ShedderGroup) GetShedder(key string) Shedder {
	shedder, _ := g.manager.Get(key, func() (io.Closer, error) {
		return nopCloser{
			Shedder: NewAdaptiveShedder(g.options...),
		}, nil
	})
	return shedder.(Shedder)
}

type nopCloser struct {
	Shedder
}

func (c nopCloser) Close() error {
	return nil
}
