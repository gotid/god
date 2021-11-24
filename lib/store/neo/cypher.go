package neo

import (
	"time"

	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/syncx"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	// 数据库慢日志阈值，用于记录慢查询和慢执行
	defaultSlowThreshold = 500 * time.Millisecond
	slowThreshold        = syncx.ForAtomicDuration(defaultSlowThreshold)
)

// SetSlowThreshold 设置慢查询时间阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}

func doRead(driver neo4j.Driver, scanner Scanner, cypher string, params ...g.Map) error {
	// 慢查询检测
	start := time.Now()

	// 执行查询
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	var param g.Map
	if len(params) > 0 {
		param = params[0]
	}
	result, err := session.Run(cypher, param)
	// fmt.Println("doRead Next:", result.Next())
	duration := time.Since(start)
	if duration > slowThreshold.Load() {
		logx.WithDuration(duration).Slowf("[Neo] doRead：慢查询 —— %s", cypher)
	} else {
		logx.WithDuration(duration).Infof("[Neo] doRead: %s", cypher)
	}

	if err != nil {
		logCypherError(cypher, err)
		return err
	}

	return scanner(result)
}

func logCypherError(cypher string, err error) {
	if err != nil {
		logx.Errorf("[Neo] %s >>> %s", err.Error(), cypher)
	}
}
