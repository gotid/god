package svc

import (
	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/config"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/model"

	"github.com/gotid/god/lib/store/sqlx"
)

type ServiceContext struct {
	c     config.Config
	Model model.ShorturlModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		c:     c,
		Model: model.NewShorturlModel(sqlx.NewMySQL(c.DataSource), c.Cache),
	}
}
