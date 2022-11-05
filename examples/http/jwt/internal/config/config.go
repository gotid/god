package config

import "github.com/gotid/god/api"

type Config struct {
	api.Config
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}
}
