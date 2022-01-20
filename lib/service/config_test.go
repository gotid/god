package service

import (
	"testing"

	"github.com/gotid/god/lib/logx"
)

func TestConfi(t *testing.T) {
	c := ServiceConf{
		Name: "foo",
		Log: logx.LogConf{
			Mode: "console",
		},
		Mode: "dev",
	}
	c.MustSetup()
}
