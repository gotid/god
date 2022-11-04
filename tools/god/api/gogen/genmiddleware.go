package gogen

import (
	_ "embed"
	"github.com/gotid/god/tools/god/api/spec"
	"github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/util/format"
	"strings"
)

//go:embed middleware.tpl
var middlewareImplementCode string

func genMiddleware(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	middlewares := getMiddleware(api)
	for _, item := range middlewares {
		middlewareFilename := strings.TrimSuffix(strings.ToLower(item), "middleware") + "_middleware"
		filename, err := format.FileNamingFormat(cfg.NamingFormat, middlewareFilename)
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(item, "Middleware") + "Middleware"
		err = genFile(fileGenConfig{
			dir:             dir,
			subDir:          middlewareDir,
			filename:        filename + ".go",
			templateName:    "contextTemplate",
			category:        category,
			templateFile:    middlewareImplementCodeFile,
			builtinTemplate: middlewareImplementCode,
			data: map[string]string{
				"name": strings.Title(name),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
