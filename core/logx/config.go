package logx

// LogConf 表示一个日志配置项。
type LogConf struct {
	// 设置服务名称，可选。在 `volume` 模式下，该名称用于生成日志文件。
	ServiceName string `json:",optional"`
	// 输出日志的模式，默认为控制台。
	// console: 将日志写入 stdout/stderr。
	// file: 将日志写入指定路径的文件。
	// volume: 该模式用于 docker，将日志写入挂载的卷。
	Mode                string `json:",default=console,options=[console,file,volume]"`
	Encoding            string `json:",default=json,options=[json,plain]"`
	TimeFormat          string `json:",optional"`
	Path                string `json:",default=logs"`
	Level               string `json:",default=info,options=[info,error,severe]"`
	Compress            bool   `json:",optional"`
	KeepDays            int    `json:",optional"`
	StackCooldownMillis int    `json:";default=100"`
	// MaxBackups 表示保留多少份日志文件。0 表示所有文件将永久保存。
	// 仅在 Rotation 为 `size` 时生效。
	// 即使 `MaxBackups` 设为 0，如果达到 `KeepDays` 限制，日志文件仍将被删除。
	MaxBackups int `json:",default=0"`
	// MaxSize 表示写入日志文件占用了多少空间。0 表示没有限制。单位是"MB"。
	// 仅在 Rotation 为 `size` 时生效。
	MaxSize int `json:",default=0"`
	// Rotation 表示日志轮换规则的类型。默认为 `daily`。
	// daily: 每日轮换。
	// size: 大小受限的轮换。
	Rotation string `json:",default=daily,options=[daily,size]"`
}
