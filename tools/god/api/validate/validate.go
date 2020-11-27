package validate

import (
	"errors"
	"fmt"

	"git.zc0901.com/go/god/tools/god/api/parser"
	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli"
)

func GoValidateApi(c *cli.Context) error {
	apiFile := c.String("api")

	if len(apiFile) == 0 {
		return errors.New("missing -api")
	}

	p, err := parser.NewParser(apiFile)
	if err != nil {
		return err
	}
	_, err = p.Parse()
	if err == nil {
		fmt.Println(aurora.Green("api format ok"))
	}
	return err
}
