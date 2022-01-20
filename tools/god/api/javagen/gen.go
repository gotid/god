package javagen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gotid/god/lib/fs"

	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/tools/god/api/parser"
	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
)

func JavaCommand(c *cli.Context) error {
	apiFile := c.String("api")
	dir := c.String("dir")
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
		return err
	}

	packetName := api.Service.Name
	if strings.HasSuffix(packetName, "-api") {
		packetName = packetName[:len(packetName)-4]
	}

	logx.Must(fs.MkdirIfNotExist(dir))
	logx.Must(genPacket(dir, packetName, api))
	logx.Must(genComponents(dir, packetName, api))

	fmt.Println(aurora.Green("完成。"))
	return nil
}
