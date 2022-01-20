package service

import (
	"testing"

	"git.zc0901.com/go/god/lib/logx"
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
