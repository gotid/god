package gogen

import (
	_ "embed"
	"fmt"
	"github.com/gotid/god/tools/god/api/spec"
	"github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/util/format"
	"github.com/gotid/god/tools/god/vars"
	"strings"
)

const (
	configFile = "config"

	jwtTemplate = ` struct {
		AccessSecret string
		AccessExpire int64
	}
`
	jwtTransTemplate = ` struct {
		Secret     string
		PrevSecret string
	}
`
)

//go:embed config.tpl
var configTemplate string

func genConfig(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, configFile)
	if err != nil {
		return err
	}

	authNames := getAuths(api)
	var auths []string
	for _, item := range authNames {
		auths = append(auths, fmt.Sprintf("%s %s", item, jwtTemplate))
	}

	jwtTransNames := getJwtTrans(api)
	var jwtTransList []string
	for _, item := range jwtTransNames {
		jwtTransList = append(jwtTransList, fmt.Sprintf("%s %s", item, jwtTransTemplate))
	}
	authImportStr := fmt.Sprintf("\"%s/api\"", vars.ProjectOpenSourceURL)

	return genFile(fileGenConfig{
		dir:             dir,
		subDir:          configDir,
		filename:        filename + ".go",
		templateName:    "configTemplate",
		category:        category,
		templateFile:    configTemplateFile,
		builtinTemplate: configTemplate,
		data: map[string]string{
			"authImport": authImportStr,
			"auth":       strings.Join(auths, "\n"),
			"jwtTrans":   strings.Join(jwtTransList, "\n"),
		},
	})
}
