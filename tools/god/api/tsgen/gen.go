package tsgen

import (
	"errors"
	"fmt"

	"github.com/gotid/god/lib/fs"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/tools/god/api/parser"
	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
)

func TsCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
	webApi := c.String("webapi")
	caller := c.String("caller")
	unwrapApi := c.Bool("unwrap")
	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}
	if len(dir) == 0 {
		return errors.New("missing -dir")
	}

	p, err := parser.NewParser(apiFile)
	if err != nil {
		return err
	}
	api, err := p.Parse()
	if err != nil {
		fmt.Println(aurora.Red("Failed"))
		return err
	}

	logx.Must(fs.MkdirIfNotExist(dir))
	logx.Must(genHandler(dir, webApi, caller, api, unwrapApi))
	logx.Must(genComponents(dir, api))

	fmt.Println(aurora.Green("完成。"))
	return nil
}
