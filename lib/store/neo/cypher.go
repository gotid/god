package neo

import (
	"time"

	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/syncx"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var (
	// 数据库慢日志阈值，用于记录慢调用和慢执行
	defaultSlowThreshold = 500 * time.Millisecond
	slowThreshold        = syncx.ForAtomicDuration(defaultSlowThreshold)
)

// SetSlowThreshold 设置慢调用时间阈值。
func SetSlowThreshold(threshold time.Duration) {
	slowThreshold.Set(threshold)
}

func doRun(driver neo4j.Driver, scanner Scanner, cypher string, params ...g.Map) error {
	// 慢调用检测
	start := time.Now()

	// 执行调用
	session := driver.NewSession(neo4j.SessionConfig{})
	defer session.Close()

	var param g.Map
	if len(params) > 0 {
		param = params[0]
	}
	result, err := session.Run(cypher, param)
	duration := time.Since(start)
	if duration > slowThreshold.Load() {
		logx.WithDuration(duration).Slowf("[Neo] doRun：慢调用 —— %s", cypher)
	} else {
		logx.WithDuration(duration).Infof("[Neo] doRun: %s", cypher)
	}

	if err != nil {
		logCypherError(cypher, err)
		return err
	}

	return scanner(result)
}

func doTxRun(tx neo4j.Transaction, scanner Scanner, cypher string, params ...g.Map) error {
	// 慢调用检测
	start := time.Now()

	// 执行调用
	var param g.Map
	if len(params) > 0 {
		param = params[0]
	}
	result, err := tx.Run(cypher, param)
	duration := time.Since(start)
	if duration > slowThreshold.Load() {
		logx.WithDuration(duration).Slowf("[Neo] doRun：慢调用 —— %s", cypher)
	} else {
		logx.WithDuration(duration).Infof("[Neo] doRun: %s", cypher)
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
