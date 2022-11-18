package logic

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gotid/god/examples/shorturl/api/internal/svc"
	"github.com/gotid/god/examples/shorturl/api/internal/types"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformer"
	"github.com/gotid/god/examples/shorturl/rpc/transformer/transformerclient"
	"github.com/stretchr/testify/assert"
)

func TestShortenLogic_Shorten(t *testing.T) {
	ast := assert.New(t)

	// Build mock and svc context
	ctl := gomock.NewController(t)
	mockTransformer := transformerclient.NewMockTransformer(ctl)

	l := NewShortenLogic(context.Background(), &svc.ServiceContext{
		Transformer: mockTransformer,
	})

	// Failed to simulate transformer Expand
	mockTransformer.EXPECT().Shorten(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("call err")).
		Times(1)

	_, err := l.Shorten(&types.ShortenRequest{})
	ast.NotNil(err)

	// Simulate transformer Expand success
	mockTransformer.EXPECT().Shorten(gomock.Any(), gomock.Any()).
		Return(&transformer.ShortenResponse{Shorten: "123"}, nil).
		Times(1)

	resp, err := l.Shorten(&types.ShortenRequest{})
	ast.Nil(err)
	ast.Equal(resp, &types.ShortenResponse{Shorten: "123"})
}
