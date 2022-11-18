package model

import (
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/sqlx"
)

var _ ShorturlModel = (*customShorturlModel)(nil)

type (
	// ShorturlModel 是一个要自定义的接口，在此添加更多方法，
	// 并在 customShorturlModel 中实现。
	ShorturlModel interface {
		shorturlModel
	}

	customShorturlModel struct {
		*defaultShorturlModel
	}
)

// NewShorturlModel 返回数据库表的模型。
func NewShorturlModel(conn sqlx.Conn, c cache.Config) ShorturlModel {
	return &customShorturlModel{
		defaultShorturlModel: newShorturlModel(conn, c),
	}
}
