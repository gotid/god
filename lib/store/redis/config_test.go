package redis

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name string
		Config
		ok bool
	}{
		{
			name: "缺少主机",
			Config: Config{
				Host: "",
				Type: NodeType,
				Pass: "",
			},
			ok: false,
		},
		{
			name: "缺少类型",
			Config: Config{
				Host: "localhost:6379",
				Type: "",
				Pass: "",
			},
			ok: false,
		},
		{
			name: "正常配置",
			Config: Config{
				Host: "localhost:6379",
				Type: NodeType,
				Pass: "",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.Config.Validate())
				assert.NotNil(t, test.Config.NewRedis())
			} else {
				assert.NotNil(t, test.Config.Validate())
			}
		})
	}
}

func TestKeyConfig(t *testing.T) {
	tests := []struct {
		name string
		KeyConfig
		ok bool
	}{
		{
			name: "缺少主机",
			KeyConfig: KeyConfig{
				Config: Config{
					Host: "",
					Type: NodeType,
					Pass: "",
				},
				Key: "foo",
			},
			ok: false,
		},
		{
			name: "缺少键名",
			KeyConfig: KeyConfig{
				Config: Config{
					Host: "localhost:6379",
					Type: NodeType,
					Pass: "",
				},
				Key: "",
			},
			ok: false,
		},
		{
			name: "正常配置",
			KeyConfig: KeyConfig{
				Config: Config{
					Host: "localhost:6379",
					Type: NodeType,
					Pass: "",
				},
				Key: "foo",
			},
			ok: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ok {
				assert.Nil(t, test.KeyConfig.Validate())
			} else {
				assert.NotNil(t, test.KeyConfig.Validate())
			}
		})
	}
}
