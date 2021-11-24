package neo

import (
	"errors"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type (
	// Scanner 是一个 neo4j.Result 的结果回调函数。
	Scanner func(result neo4j.Result) error
)

var (
	ErrNoScanner   = errors.New("请指定结果扫描函数")
	ErrNotSettable = errors.New("扫描目标不可设置")
)
