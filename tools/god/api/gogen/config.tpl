package config

import {{.authImport}}

type Config struct {
	api.Config
	{{.auth}}
	{{.jwtTrans}}
}
