package model

import (
	"github.com/gotid/god/tools/god/model/sql/command"
	"github.com/spf13/cobra"
)

var (
	// Cmd 描述了一个 model 命令。
	Cmd = &cobra.Command{
		Use:   "model",
		Short: "生成 model 模型",
	}

	mysqlCmd = &cobra.Command{
		Use:   "mysql",
		Short: "生成 mysql 模型",
	}

	dsnCmd = &cobra.Command{
		Use:   "dsn",
		Short: "从数据库连接生成模型",
		RunE:  command.MySqlDSN,
	}

	ddlCmd = &cobra.Command{
		Use:   "dsn",
		Short: "从数据库脚本生成模型",
		RunE:  command.MySqlDDL,
	}
)

func init() {
	mysqlCmd.AddCommand(dsnCmd)
	mysqlCmd.AddCommand(ddlCmd)

	Cmd.AddCommand(mysqlCmd)
}
