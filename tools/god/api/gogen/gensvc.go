package gogen

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/gotid/god/tools/god/api/spec"
	"github.com/gotid/god/tools/god/api/util"
	"github.com/gotid/god/tools/god/config"
	ctlutil "github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/vars"
)

const (
	contextFilename = "service_context"
	contextTemplate = `package svc

import (
	{{.configImport}}
)

type ServiceContext struct {
	Config {{.config}}
	{{.middleware}}
}

func NewServiceContext(c {{.config}}) *ServiceContext {
	return &ServiceContext{
		Config: c, 
		{{.middlewareAssignment}}
	}
}

`
)

func genServiceContext(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, contextFilename)
	if err != nil {
		return err
	}

	fp, created, err := util.MaybeCreateFile(dir, contextDir, filename+".go")
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	authNames := getAuths(api)
	var auths []string
	for _, item := range authNames {
		auths = append(auths, fmt.Sprintf("%s config.AuthConfig", item))
	}

	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	text, err := ctlutil.LoadTemplate(category, contextTemplateFile, contextTemplate)
	if err != nil {
		return err
	}

	var middlewareStr string
	var middlewareAssignment string
	middlewares := getMiddleware(api)

	for _, item := range middlewares {
		middlewareStr += fmt.Sprintf("%s api.Middleware\n", item)
		name := strings.TrimSuffix(item, "Middleware") + "Middleware"
		middlewareAssignment += fmt.Sprintf("%s: %s,\n", item, fmt.Sprintf("middleware.New%s().%s", strings.Title(name), "Handle"))
	}

	configImport := "\"" + ctlutil.JoinPackages(parentPkg, configDir) + "\""
	if len(middlewareStr) > 0 {
		configImport += "\n\t\"" + ctlutil.JoinPackages(parentPkg, middlewareDir) + "\""
		configImport += fmt.Sprintf("\n\t\"%s/api\"", vars.ProjectOpenSourceUrl)
	}

	t := template.Must(template.New("contextTemplate").Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"configImport":         configImport,
		"config":               "config.Config",
		"middleware":           middlewareStr,
		"middlewareAssignment": middlewareAssignment,
	})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}
