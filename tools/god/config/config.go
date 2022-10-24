package config

import (
	"errors"
	"strings"
)

const DefaultFormat = "user_login"

// Config 定义了文件命名的样式。
type Config struct {
	// 用于定义生成的文件命名样式。
	// 如蛇式命名：user_login，或小驼峰命名：userLogin。
	// 理论上，也可指定分割符：user#login，但需遵循操作系统的文件命名规范。
	// 注意：NamingFormat 基于蛇式或驼峰。
	NamingFormat string `yaml:"namingFormat"`
}

// NewConfig 返回一个新的 Config。
func NewConfig(format string) (*Config, error) {
	if len(format) == 0 {
		format = DefaultFormat
	}

	cfg := &Config{NamingFormat: format}
	err := validate(cfg)
	return cfg, err
}

func validate(cfg *Config) error {
	if len(strings.TrimSpace(cfg.NamingFormat)) == 0 {
		return errors.New("缺少配置项 - namingFormat")
	}

	return nil
}
