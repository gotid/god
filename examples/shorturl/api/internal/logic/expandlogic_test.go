package logic

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/gotid/god/examples/shorturl/api/internal/svc"
	"github.com/gotid/god/examples/shorturl/api/internal/types"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformerclient"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExpandLogic_Expand(t *testing.T) {
	ast := assert.New(t)

	// Build mock and svc context
	ctl := gomock.NewController(t)
	mockTransformer := transformerclient.NewMockTransformer(ctl)

	l := NewExpandLogic(context.Background(), &svc.ServiceContext{
		Transformer: mockTransformer,
	})

	// Failed to simulate transformer Expand
	mockTransformer.EXPECT().Expand(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("call err")).
		Times(1)

	_, err := l.Expand(&types.ExpandRequest{})
	ast.NotNil(err)

	// Simulate transformer Expand success
	mockTransformer.EXPECT().Expand(gomock.Any(), gomock.Any()).
		Return(&transformerclient.ExpandResponse{Url: "url"}, nil).
		Times(1)

	resp, err := l.Expand(&types.ExpandRequest{})
	ast.Nil(err)
	ast.Equal(resp, &types.ExpandResponse{Url: "url"})
}
