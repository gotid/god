package gogen

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"git.zc0901.com/go/god/lib/collection"
	"git.zc0901.com/go/god/tools/god/api/spec"
	apiutil "git.zc0901.com/go/god/tools/god/api/util"
	"git.zc0901.com/go/god/tools/god/config"
	"git.zc0901.com/go/god/tools/god/util"
	"git.zc0901.com/go/god/tools/god/util/format"
	"git.zc0901.com/go/god/tools/god/vars"
)

const (
	routesFilename = "routes"
	routesTemplate = `// Code generated by god. DO NOT EDIT.
package handler

import (
	"net/http"

	{{.importPackages}}
)

func RegisterHandlers(engine *api.Server, serverCtx *svc.ServiceContext) {
	{{.routesAdditions}}
}
`
	routesAdditionTemplate = `
	engine.AddRoutes(
		{{.routes}} {{.jwt}}{{.signature}}
	)
`
)

var mapping = map[string]string{
	"delete": "http.MethodDelete",
	"get":    "http.MethodGet",
	"head":   "http.MethodHead",
	"post":   "http.MethodPost",
	"put":    "http.MethodPut",
	"patch":  "http.MethodPatch",
}

type (
	group struct {
		routes           []route
		jwtEnabled       bool
		signatureEnabled bool
		authName         string
		middlewares      []string
	}
	route struct {
		method  string
		path    string
		handler string
	}
)

func genRoutes(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	var builder strings.Builder
	groups, err := getRoutes(api)
	if err != nil {
		return err
	}

	gt := template.Must(template.New("groupTemplate").Parse(routesAdditionTemplate))
	for _, g := range groups {
		var gbuilder strings.Builder
		gbuilder.WriteString("[]api.Route{")
		for _, r := range g.routes {
			fmt.Fprintf(&gbuilder, `
		{
			Method:  %s,
			Path:    "%s",
			Handler: %s,
		},`,
				r.method, r.path, r.handler)
		}

		var jwt string
		if g.jwtEnabled {
			jwt = fmt.Sprintf("\n api.WithJwt(serverCtx.Config.%s.AccessSecret),", g.authName)
		}
		var signature string
		if g.signatureEnabled {
			signature = fmt.Sprintf("\n api.WithSignature(serverCtx.Config.%s.Signature),", g.authName)
		}

		var routes string
		if len(g.middlewares) > 0 {
			gbuilder.WriteString("\n}...,")
			params := g.middlewares
			for i := range params {
				params[i] = "serverCtx." + params[i]
			}
			middlewareStr := strings.Join(params, ", ")
			routes = fmt.Sprintf("api.WithMiddlewares(\n[]api.Middleware{ %s }, \n %s \n),",
				middlewareStr, strings.TrimSpace(gbuilder.String()))
		} else {
			gbuilder.WriteString("\n},")
			routes = strings.TrimSpace(gbuilder.String())
		}

		if err := gt.Execute(&builder, map[string]string{
			"routes":    routes,
			"jwt":       jwt,
			"signature": signature,
		}); err != nil {
			return err
		}
	}

	parentPkg, err := getParentPackage(dir)
	if err != nil {
		return err
	}

	routeFilename, err := format.FileNamingFormat(cfg.NamingFormat, routesFilename)
	if err != nil {
		return err
	}
	routeFilename = routeFilename + ".go"

	filename := path.Join(dir, handlerDir, routeFilename)
	os.Remove(filename)

	fp, created, err := apiutil.MaybeCreateFile(dir, handlerDir, routeFilename)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	t := template.Must(template.New("routesTemplate").Parse(routesTemplate))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, map[string]string{
		"importPackages":  genRouteImports(parentPkg, api),
		"routesAdditions": strings.TrimSpace(builder.String()),
	})
	if err != nil {
		return err
	}

	formatCode := formatCode(buffer.String())
	_, err = fp.WriteString(formatCode)
	return err
}

func genRouteImports(parentPkg string, api *spec.ApiSpec) string {
	importSet := collection.NewSet()
	importSet.AddStr(fmt.Sprintf("\"%s\"", util.JoinPackages(parentPkg, contextDir)))
	for _, group := range api.Service.Groups {
		for _, route := range group.Routes {
			folder, ok := apiutil.GetAnnotationValue(route.Annotations, "server", groupProperty)
			if !ok {
				folder, ok = apiutil.GetAnnotationValue(group.Annotations, "server", groupProperty)
				if !ok {
					continue
				}
			}
			importSet.AddStr(fmt.Sprintf("%s \"%s\"", toPrefix(folder), util.JoinPackages(parentPkg, handlerDir, folder)))
		}
	}
	imports := importSet.KeysStr()
	sort.Strings(imports)
	projectSection := strings.Join(imports, "\n\t")
	depSection := fmt.Sprintf("\"%s/api\"", vars.ProjectOpenSourceUrl)
	return fmt.Sprintf("%s\n\n\t%s", projectSection, depSection)
}

func getRoutes(api *spec.ApiSpec) ([]group, error) {
	var routes []group

	for _, g := range api.Service.Groups {
		prefix, _ := apiutil.GetAnnotationValue(g.Annotations, "server", prefixProperty)

		var groupedRoutes group
		for _, r := range g.Routes {
			handler := getHandlerName(r)
			handler = handler + "(serverCtx)"
			folder, ok := apiutil.GetAnnotationValue(r.Annotations, "server", groupProperty)
			if ok {
				handler = toPrefix(folder) + "." + strings.ToUpper(handler[:1]) + handler[1:]
			} else {
				folder, ok = apiutil.GetAnnotationValue(g.Annotations, "server", groupProperty)
				if ok {
					handler = toPrefix(folder) + "." + strings.ToUpper(handler[:1]) + handler[1:]
				}
			}
			routePath := r.Path
			if prefix != "" {
				routePath = prefix + routePath
			}
			groupedRoutes.routes = append(groupedRoutes.routes, route{
				method:  mapping[r.Method],
				path:    routePath,
				handler: handler,
			})
		}

		if value, ok := apiutil.GetAnnotationValue(g.Annotations, "server", "jwt"); ok {
			groupedRoutes.authName = value
			groupedRoutes.jwtEnabled = true
		}
		if value, ok := apiutil.GetAnnotationValue(g.Annotations, "server", "middleware"); ok {
			for _, item := range strings.Split(value, ",") {
				groupedRoutes.middlewares = append(groupedRoutes.middlewares, item)
			}
		}
		routes = append(routes, groupedRoutes)
	}

	return routes, nil
}

func toPrefix(folder string) string {
	return strings.ReplaceAll(folder, "/", "")
}
