package neo

import (
	"fmt"
	"time"

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

func doRun(ctx Context, scanner Scanner, cypher string) error {
	// 慢调用检测
	start := time.Now()

	// 执行调用
	var result neo4j.Result
	var err error
	if ctx.Tx != nil {
		result, err = ctx.Tx.Run(cypher, ctx.Params)
	} else {
		session := ctx.Driver.NewSession(neo4j.SessionConfig{})
		defer session.Close()
		result, err = session.Run(cypher, ctx.Params)
	}
	ctx.Params = nil

	// 慢调用记录
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

	if scanner != nil {
		return scanner(result)
	}

	return nil
}

func doTx(tx neo4j.Transaction, fn TransactFn) (err error) {
	defer func() {
		if p := recover(); p != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务来自 %v, 回滚失败: %v", p, e)
			} else {
				err = fmt.Errorf("事务回滚成功，源于错误: %v", p)
			}
		} else if err != nil {
			if e := tx.Rollback(); e != nil {
				err = fmt.Errorf("事务失败: %s, 回滚失败: %s", err, e)
			}
		} else {
			err = tx.Commit()
		}
	}()

	return fn(tx)
}

func logCypherError(cypher string, err error) {
	if err != nil {
		logx.Errorf("[Neo] %s >>> %s", err.Error(), cypher)
	}
}
