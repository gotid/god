{{blue "用法："}}{{if .Runnable}}
  {{green .UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{green .CommandPath}} [命令]{{end}}{{if gt (len .Aliases) 0}}

{{blue "别名："}}
  {{green .NameAndAliases}}{{end}}{{if .HasExample}}

{{blue "示例："}}
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{blue "可用命令："}}{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rPadX .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{blue "标示："}}
{{green .LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{blue "全局标示："}}
{{green .InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{blue "其他帮助主题："}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rPadX .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{green .CommandPath}} [command] --help" for more information about a command.{{end}}
