package parser

import "github.com/gotid/god/tools/god/api/spec"

type state interface {
	process(api *spec.ApiSpec) (state, error)
}
