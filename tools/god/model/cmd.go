package model

import (
	"github.com/gotid/god/tools/god/model/sql/command"
	"github.com/spf13/cobra"
)

var (
	// Cmd 描述了一个 mysql model 命令。
	Cmd = &cobra.Command{
		Use:   "mysql",
		Short: "生成 mysql 模型",
	}

	ddlCmd = &cobra.Command{
		Use:   "ddl",
		Short: "从数据库脚本生成模型",
		RunE:  command.MySqlDDL,
	}

	dsnCmd = &cobra.Command{
		Use:   "dsn",
		Short: "从数据库连接生成模型",
		RunE:  command.MySqlDSN,
	}
)

func init() {
	ddlCmd.Flags().StringVarP(&command.VarStringSrc, "src", "s", "", "ddl 脚本路径或通配符样式")
	ddlCmd.Flags().StringVarP(&command.VarStringDir, "dir", "d", "", "目标目录")
	ddlCmd.Flags().StringVar(&command.VarStringStyle, "style", "", "文件命名样式，详见 [https://github.com/gotid/god/blob/master/tools/god/config/readme.md]")
	ddlCmd.Flags().BoolVarP(&command.VarBoolCache, "cache", "c", false, "生成缓存代码[可选]")
	ddlCmd.Flags().BoolVar(&command.VarBoolIdea, "idea", false, "用于 idea 插件[可选]")
	ddlCmd.Flags().StringVar(&command.VarStringDatabase, "database", "", "数据库名称[可选]")
	ddlCmd.Flags().StringVar(&command.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")
	ddlCmd.Flags().StringVar(&command.VarStringRemote, "remote", "", "远程 git 模板仓库，优先级高于 home\n\t模板目录要与 https://github.com/gotid/god-template 保持一致")
	ddlCmd.Flags().StringVar(&command.VarStringBranch, "branch", "", "远程仓库分值，与 --remote 配合使用")

	dsnCmd.Flags().StringVar(&command.VarStringURL, "url", "", `数据库的数据源，例如 "root:password@tcp(127.0.0.1:3306)/database"`)
	dsnCmd.Flags().StringSliceVarP(&command.VarStringSliceTable, "table", "t", nil, "数据库中的表或表通配符")
	dsnCmd.Flags().BoolVarP(&command.VarBoolCache, "cache", "c", false, "生成缓存代码[可选]")
	dsnCmd.Flags().StringVarP(&command.VarStringDir, "dir", "d", "", "目标目录")
	dsnCmd.Flags().StringVar(&command.VarStringStyle, "style", "", "文件命名样式，详见 [https://github.com/gotid/god/blob/master/tools/god/config/readme.md]")
	dsnCmd.Flags().BoolVar(&command.VarBoolIdea, "idea", false, "用于 idea 插件[可选]")
	dsnCmd.Flags().StringVar(&command.VarStringHome, "home", "", "god 模板主目录，--remote 优先级高于 --home")
	dsnCmd.Flags().StringVar(&command.VarStringRemote, "remote", "", "远程 git 模板仓库，优先级高于 home\n\t模板目录要与 https://github.com/gotid/god-template 保持一致")
	dsnCmd.Flags().StringVar(&command.VarStringBranch, "branch", "", "远程仓库分值，与 --remote 配合使用")

	Cmd.AddCommand(ddlCmd)
	Cmd.AddCommand(dsnCmd)
}
