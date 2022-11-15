package config

import "github.com/gotid/god/api"

type Config struct {
	api.Config
	Path string `json:",default=."`
}
