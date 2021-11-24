package neo

import (
	"time"
)

// Config 是一个连接了 neo4j 的配置项。
type Config struct {
	Target, Username, Password string        // neo4j 驱动连接
	Limit                      int           // neo4j 返回条数限制
	Timeout                    time.Duration // neo4j 执行超时时长
}

// NewConfig 返回一个新的连接了 neo4j 的配置项。
func NewConfig(target, username, password string,
	limit int, timeout time.Duration) *Config {
	if limit < 1 {
		limit = 10
	}
	if target == "" ||
		username == "" ||
		password == "" {
		panic("neo4j 驱动连接信息不能为空！")
	}
	return &Config{
		Target:   target,
		Username: username,
		Password: password,
		Limit:    limit,
		Timeout:  timeout,
	}
}
