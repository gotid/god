package gogen

import (
	_ "embed"
	"fmt"
	"github.com/gotid/god/tools/god/api/spec"
	"github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/util/pathx"
	"github.com/gotid/god/tools/god/vars"
	"strings"
)

const contextFilename = "service_context"

//go:embed svc.tpl
var contextTemplate string

func genServiceContext(dir, rootPkg string, cfg *config.Config, api *spec.ApiSpec) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, contextFilename)
	if err != nil {
		return err
	}

	var middlewareStr string
	var middlewareAssignment string
	middlewares := getMiddleware(api)

	for _, item := range middlewares {
		middlewareStr += fmt.Sprintf("%s api.Middleware\n", item)
		name := strings.TrimSuffix(item, "Middleware") + "Middleware"
		middlewareAssignment += fmt.Sprintf("%s: %s,\n", item,
			fmt.Sprintf("middleware.New%s().%s", strings.Title(name), "Handle"))
	}

	configImport := "\"" + pathx.JoinPackages(rootPkg, configDir) + "\""
	if len(middlewareStr) > 0 {
		configImport += "\n\t\"" + pathx.JoinPackages(rootPkg, middlewareDir) + "\""
		configImport += fmt.Sprintf("\n\t\"%s/api\"", vars.ProjectOpenSourceURL)
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subDir:          contextDir,
		filename:        filename + ".go",
		templateName:    "contextTemplate",
		category:        category,
		templateFile:    contextTemplateFile,
		builtinTemplate: contextTemplate,
		data: map[string]string{
			"configImport":         configImport,
			"config":               "config.Config",
			"middleware":           middlewareStr,
			"middlewareAssignment": middlewareAssignment,
		},
	})
}
