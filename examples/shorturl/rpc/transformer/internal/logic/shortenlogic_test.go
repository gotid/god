package logic

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/svc"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/model"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShortenLogic_Shorten(t *testing.T) {
	// 构建模拟模型和服务上下文
	ctl := gomock.NewController(t)
	shortModel := model.NewMockshorturlModel(ctl)
	svcCtx := &svc.ServiceContext{
		Model: shortModel,
	}

	// 构建短网址逻辑
	logic := NewShortenLogic(context.Background(), svcCtx)

	// 模拟模型插入失败
	shortModel.EXPECT().Insert(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("插入失败")).
		Times(1)
	_, err := logic.Shorten(&transformer.ShortenRequest{Url: "testUrl"})
	assert.NotNil(t, err)

	// 模拟模型插入成功
	shortModel.EXPECT().Insert(gomock.Any(), gomock.Any()).
		Return(nil, nil).
		Times(1)
	resp, err := logic.Shorten(&transformer.ShortenRequest{Url: "testUrl"})
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.Shorten)
	fmt.Println(resp.Shorten)
}
