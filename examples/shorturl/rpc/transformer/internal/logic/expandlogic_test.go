package logic

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/gotid/god/examples/shorturl/rpc/transformer/internal/svc"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/model"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"
)

func TestExpandLogic_Expand(t *testing.T) {
	// 构建模拟模型和服务上下文
	ctl := gomock.NewController(t)
	shortModel := model.NewMockshorturlModel(ctl)
	svcCtx := &svc.ServiceContext{
		Model: shortModel,
	}

	// 构建短网址逻辑
	logic := NewExpandLogic(context.Background(), svcCtx)

	// 模拟模型单查失败
	shortModel.EXPECT().FindOne(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("单查失败")).
		Times(1)
	_, err := logic.Expand(&transformer.ExpandRequest{})
	assert.NotNil(t, err)

	// 模拟模型单查成功
	shortModel.EXPECT().FindOne(gomock.Any(), gomock.Any()).
		Return(&model.Shorturl{
			Shorten: "testShorten",
			Url:     "testUrl",
		}, nil).
		Times(1)
	resp, err := logic.Expand(&transformer.ExpandRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.Url)
}
