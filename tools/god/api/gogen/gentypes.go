package gogen

import (
	_ "embed"
	"fmt"
	"github.com/gotid/god/tools/god/api/spec"
	apiUtil "github.com/gotid/god/tools/god/api/util"
	"github.com/gotid/god/tools/god/config"
	"github.com/gotid/god/tools/god/util"
	"github.com/gotid/god/tools/god/util/format"
	"io"
	"os"
	"path"
	"strings"
)

const typesFile = "types"

//go:embed types.tpl
var typesTemplate string

// BuildTypes gen types to string
func BuildTypes(types []spec.Type) (string, error) {
	var builder strings.Builder
	first := true
	for _, tp := range types {
		if first {
			first = false
		} else {
			builder.WriteString("\n\n")
		}
		if err := writeType(&builder, tp); err != nil {
			return "", apiUtil.WrapErr(err, "类型 "+tp.Name()+" 生成错误")
		}
	}

	return builder.String(), nil
}

func genTypes(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	val, err := BuildTypes(api.Types)
	if err != nil {
		return err
	}

	typeFilename, err := format.FileNamingFormat(cfg.NamingFormat, typesFile)
	if err != nil {
		return err
	}

	typeFilename = typeFilename + ".go"
	filename := path.Join(dir, typesDir, typeFilename)
	os.Remove(filename)

	return genFile(fileGenConfig{
		dir:             dir,
		subDir:          typesDir,
		filename:        typeFilename,
		templateName:    "typesTemplate",
		category:        category,
		templateFile:    typesTemplateFile,
		builtinTemplate: typesTemplate,
		data: map[string]interface{}{
			"types":        val,
			"containsTime": false,
		},
	})
}

func writeType(
	writer io.Writer,
	tp spec.Type,
) error {
	structType, ok := tp.(spec.DefineStruct)
	if !ok {
		return fmt.Errorf("不支持的结构体类型：%s", tp.Name())
	}

	fmt.Fprintf(writer, "type %s struct {\n", util.Title(tp.Name()))
	for _, member := range structType.Members {
		if member.IsInline {
			if _, err := fmt.Fprintf(writer, "%s\n", strings.Title(member.Type.Name())); err != nil {
				return err
			}

			continue
		}

		if err := writeProperty(writer, member.Name, member.Tag, member.GetComment(), member.Type, 1); err != nil {
			return err
		}
	}
	fmt.Fprintf(writer, "}")
	return nil
}
