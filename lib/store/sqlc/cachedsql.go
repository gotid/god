package sqlc

import (
	"context"
	"database/sql"
	"github.com/gotid/god/lib/store/cache"
	"github.com/gotid/god/lib/store/redis"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/lib/syncx"
	"time"
)

// 防止缓存同时过期造成雪崩。
const cacheSafeGapBetweenIndexAndPrimary = 5 * time.Second

var (
	// ErrNotFound 是 sqlx.ErrNotFound 的别名。
	ErrNotFound = sqlx.ErrNotFound

	singleFlights = syncx.NewSingleFlight()
	stats         = cache.NewStat("sqlc")
)

type (
	// CachedConn 是一个带缓存能力的数据库连接。
	CachedConn struct {
		db    sqlx.Conn
		cache cache.Cache
	}

	// ExecFn 定义 sql 执行方法。
	ExecFn func(conn sqlx.Conn) (sql.Result, error)
	// ExecCtxFn 定义 sql 执行方法。
	ExecCtxFn func(ctx context.Context, conn sqlx.Conn) (sql.Result, error)
	// QueryFn 定义 sql 查询方法。
	QueryFn func(conn sqlx.Conn, v interface{}) error
	// QueryCtxFn 定义 sql 查询方法。
	QueryCtxFn func(ctx context.Context, conn sqlx.Conn, v interface{}) error
	// IndexQueryFn 定义基于唯一索引的 sql 查询方法。
	IndexQueryFn func(conn sqlx.Conn, v interface{}) (interface{}, error)
	// IndexQueryCtxFn 定义基于唯一索引的 sql 查询方法。
	IndexQueryCtxFn func(ctx context.Context, conn sqlx.Conn, v interface{}) (interface{}, error)
	// PrimaryQueryFn 定义基于主键的 sql 查询方法。
	PrimaryQueryFn func(conn sqlx.Conn, v, primary interface{}) error
	// PrimaryQueryCtxFn 定义基于主键的 sql 查询方法。
	PrimaryQueryCtxFn func(ctx context.Context, conn sqlx.Conn, v, primary interface{}) error
)

// NewConn 返回一个给定 redis 集群的数据库连接 CachedConn。
func NewConn(db sqlx.Conn, config cache.Config, opts ...cache.Option) CachedConn {
	c := cache.New(config, singleFlights, stats, sql.ErrNoRows, opts...)
	return NewConnWithCache(db, c)

}

// NewNodeConn 换回一个给定 redis 节点的数据库连接 CachedConn。
func NewNodeConn(db sqlx.Conn, rds *redis.Redis, opts ...cache.Option) CachedConn {
	c := cache.NewNode(rds, singleFlights, stats, sql.ErrNoRows, opts...)
	return NewConnWithCache(db, c)
}

// NewConnWithCache 返回一个给定自定义缓存的数据库连接 CachedConn。
func NewConnWithCache(db sqlx.Conn, c cache.Cache) CachedConn {
	return CachedConn{
		db:    db,
		cache: c,
	}
}

// DelCache 删除给定键的缓存。
func (cc CachedConn) DelCache(keys ...string) error {
	return cc.DelCacheCtx(context.Background(), keys...)
}

// DelCacheCtx 删除给定键的缓存。
func (cc CachedConn) DelCacheCtx(ctx context.Context, keys ...string) error {
	return cc.cache.DelCtx(ctx, keys...)
}

// GetCache 获取给定键的缓存并解编组至变量 v。
func (cc CachedConn) GetCache(key string, v interface{}) error {
	return cc.GetCacheCtx(context.Background(), key, v)
}

// GetCacheCtx 获取给定键的缓存并解编组至变量 v。
func (cc CachedConn) GetCacheCtx(ctx context.Context, key string, v interface{}) error {
	return cc.cache.GetCtx(ctx, key, v)
}

// Exec 对给定键执行给定函数，并返回执行结果。
func (cc CachedConn) Exec(exec ExecFn, keys ...string) (sql.Result, error) {
	execCtx := func(_ context.Context, conn sqlx.Conn) (sql.Result, error) {
		return exec(conn)
	}

	return cc.ExecCtx(context.Background(), execCtx, keys...)
}

// ExecCtx 对给定键执行给定函数（且尝试删除缓存），并返回执行结果。
func (cc CachedConn) ExecCtx(ctx context.Context, exec ExecCtxFn, keys ...string) (sql.Result, error) {
	res, err := exec(ctx, cc.db)
	if err != nil {
		return nil, err
	}

	if err := cc.DelCacheCtx(ctx, keys...); err != nil {
		return nil, err
	}

	return res, nil
}

// ExecNoCache 运行给定的 sql 语句，但不删除缓存。
func (cc CachedConn) ExecNoCache(query string, args ...interface{}) (sql.Result, error) {
	return cc.ExecNoCacheCtx(context.Background(), query, args...)
}

// ExecNoCacheCtx 运行给定的 sql 语句，但不删除缓存。
func (cc CachedConn) ExecNoCacheCtx(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return cc.db.ExecCtx(ctx, query, args...)
}

// QueryRow 查询缓存键 key 的值到变量 v，若无缓存则使用 query 函数进行查询并编组至缓存建。
func (cc CachedConn) QueryRow(v interface{}, key string, query QueryFn) error {
	queryCtx := func(_ context.Context, conn sqlx.Conn, v interface{}) error {
		return query(conn, v)
	}

	return cc.QueryRowCtx(context.Background(), v, key, queryCtx)
}

// QueryRowCtx 查询缓存键 key 的值到变量 v，若无缓存则使用 query 函数进行查询并编组至缓存建。
func (cc CachedConn) QueryRowCtx(ctx context.Context, v interface{}, key string, query QueryCtxFn) error {
	return cc.cache.TakeCtx(ctx, v, key, func(val interface{}) error {
		return query(ctx, cc.db, val)
	})
}

// QueryRowIndex 查询缓存建 key 的值到变量 v。
func (cc CachedConn) QueryRowIndex(v interface{}, key string, keyer func(primary interface{}) string,
	indexQuery IndexQueryFn, primaryQuery PrimaryQueryFn) error {
	indexQueryCtx := func(_ context.Context, conn sqlx.Conn, v interface{}) (interface{}, error) {
		return indexQuery(conn, v)
	}
	primaryQueryCtx := func(_ context.Context, conn sqlx.Conn, v, primary interface{}) error {
		return primaryQuery(conn, v, primary)
	}

	return cc.QueryRowIndexCtx(context.Background(), v, key, keyer, indexQueryCtx, primaryQueryCtx)
}

// QueryRowIndexCtx 查询缓存建 key 的值到变量 v。
func (cc CachedConn) QueryRowIndexCtx(ctx context.Context, v interface{}, key string,
	keyer func(primary interface{}) string, indexQuery IndexQueryCtxFn, primaryQuery PrimaryQueryCtxFn) error {
	var primaryKey interface{}
	var found bool

	err := cc.cache.TakeWithExpireCtx(ctx, &primaryKey, key,
		func(val interface{}, expire time.Duration) (err error) {
			primaryKey, err = indexQuery(ctx, cc.db, v)
			if err != nil {
				return
			}

			found = true
			return cc.cache.SetWithExpireCtx(ctx, keyer(primaryKey), v, expire+cacheSafeGapBetweenIndexAndPrimary)
		})
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	return cc.cache.TakeCtx(ctx, v, keyer(primaryKey), func(val interface{}) error {
		return primaryQuery(ctx, cc.db, v, primaryKey)
	})
}

// QueryRowNoCache 查询给定 sql 语句结果至变量 v。
// 未使用缓存，可能导致数据不一致。
func (cc CachedConn) QueryRowNoCache(v interface{}, query string, args ...interface{}) error {
	return cc.QueryRowNoCacheCtx(context.Background(), v, query, args...)
}

// QueryRowNoCacheCtx 查询给定 sql 语句结果至变量 v。
// 未使用缓存，可能导致数据不一致。
func (cc CachedConn) QueryRowNoCacheCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error {
	return cc.db.QueryRowCtx(ctx, v, query, args...)
}

// QueryRowsNoCache 查询给定 sql 语句结果至变量 v。
// 未使用缓存，可能导致数据不一致。
func (cc CachedConn) QueryRowsNoCache(v interface{}, query string, args ...interface{}) error {
	return cc.QueryRowsNoCacheCtx(context.Background(), v, query, args...)
}

// QueryRowsNoCacheCtx 查询给定 sql 语句结果至变量 v。
// 未使用缓存，可能导致数据不一致。
func (cc CachedConn) QueryRowsNoCacheCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error {
	return cc.db.QueryRowsCtx(ctx, v, query, args...)
}

// SetCache 设置键值对缓存，并将其存活时间设置为节点指定的时间。
func (cc CachedConn) SetCache(key string, val interface{}) error {
	return cc.SetCacheCtx(context.Background(), key, val)
}

// SetCacheCtx 设置键值对缓存，并将其存活时间设置为节点指定的时间。
func (cc CachedConn) SetCacheCtx(ctx context.Context, key string, val interface{}) error {
	return cc.cache.SetCtx(ctx, key, val)
}

// Transact 在事务模式中运行给定函数。
func (cc CachedConn) Transact(fn func(sqlx.Session) error) error {
	fnCtx := func(_ context.Context, session sqlx.Session) error {
		return fn(session)
	}

	return cc.TransactCtx(context.Background(), fnCtx)
}

// TransactCtx 在事务模式中运行给定函数。
func (cc CachedConn) TransactCtx(ctx context.Context, fn func(context.Context, sqlx.Session) error) error {
	return cc.db.TransactCtx(ctx, fn)
}
