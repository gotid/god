package service

import (
	"github.com/gotid/god/lib/logx"
	"testing"
)

func TestConfig(t *testing.T) {
	config := Config{
		Name: "foo",
		Mode: "dev",
		Log: logx.Config{
			Mode: "console",
		},
	}
	config.MustSetup()
}
