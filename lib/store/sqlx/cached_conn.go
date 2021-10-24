package sqlx

import (
	"database/sql"
	"time"

	"git.zc0901.com/go/god/lib/store/cache"
	"git.zc0901.com/go/god/lib/store/redis"
	"git.zc0901.com/go/god/lib/syncx"
)

// 索引建和主键缓存之间的安全时间间隔
const safeGapBetweenIndexAndPrimary = 5 * time.Second

var (
	exclusiveCalls = syncx.NewSingleFlight()
	cacheStat      = cache.NewCacheStat("cached_conn")
)

type (
	CachedConn struct {
		conn  Conn
		cache cache.Cache
	}

	ExecFn  func(conn Conn) (sql.Result, error)     // 常规的写库函数
	QueryFn func(conn Conn, dest interface{}) error // 常规的读库函数

	GetKeyOfPKFn   func(pk interface{}) string                                   // 取主键的缓存键
	IndexQueryFn   func(conn Conn, dest interface{}) (pk interface{}, err error) // 按索引查行结果
	PrimaryQueryFn func(conn Conn, dest, pk interface{}) error                   // 按主键查行结果
)

func NewCachedConn(conn Conn, rds *redis.Redis, opts ...cache.Option) CachedConn {
	return CachedConn{
		conn:  conn,
		cache: cache.NewCacheNode(rds, exclusiveCalls, cacheStat, ErrNotFound, opts...),
	}
}

func NewCachedConnWithCluster(conn Conn, c cache.ClusterConf, opts ...cache.Option) CachedConn {
	return CachedConn{
		conn:  conn,
		cache: cache.NewCacheCluster(c, exclusiveCalls, cacheStat, sql.ErrNoRows, opts...),
	}
}

func (cc CachedConn) DelCache(keys ...string) error {
	return cc.cache.Del(keys...)
}

func (cc CachedConn) GetCache(key string, dest interface{}) error {
	return cc.cache.Get(key, dest)
}

func (cc CachedConn) MGetCache(keys []string, dest []interface{}) error {
	return cc.cache.MGet(keys, dest)
}

func (cc CachedConn) SetCache(key string, value interface{}) error {
	return cc.cache.Set(key, value)
}

func (cc CachedConn) SetBit(key string, offset int64, value int) error {
	return cc.cache.SetBit(key, offset, value)
}

func (cc CachedConn) SetBits(key string, offset []int64) error {
	return cc.cache.SetBits(key, offset)
}

func (cc CachedConn) GetBit(key string, offset int64) (int, error) {
	return cc.cache.GetBit(key, offset)
}

func (cc CachedConn) GetBits(key string, offset []int64) (map[int64]bool, error) {
	return cc.cache.GetBits(key, offset)
}

// Exec 执行增、删、改，并清空 keys 对应的缓存
func (cc CachedConn) Exec(exec ExecFn, keys ...string) (sql.Result, error) {
	result, err := exec(cc.conn)
	if err != nil {
		return nil, err
	}

	if err := cc.DelCache(keys...); err != nil {
		return nil, err
	}

	return result, nil
}

// ExecNoCache 无缓存执行增、删、改
func (cc CachedConn) ExecNoCache(query string, args ...interface{}) (sql.Result, error) {
	return cc.conn.Exec(query, args...)
}

// Query 先按 key 从缓存拿，拿不到则查库、写缓存并返回新值
func (cc CachedConn) Query(dest interface{}, key string, query QueryFn) error {
	return cc.cache.Take(dest, key, func(dbValue interface{}) error {
		return query(cc.conn, dbValue)
	})
}

// QueryNoCache 无缓存查询，直接读库
func (cc CachedConn) QueryNoCache(dest interface{}, query string, args ...interface{}) error {
	return cc.conn.Query(dest, query, args...)
}

func (cc CachedConn) Transact(fn func(tx TxSession) error) error {
	return cc.conn.Transact(fn)
}

func (cc CachedConn) QueryIndex(dest interface{}, indexKey string, getKeyOfPK GetKeyOfPKFn,
	indexQuery IndexQueryFn, primaryQuery PrimaryQueryFn) error {
	var id interface{}
	var found bool

	// 转变主键类型为 int32
	getKeyOfPK = toInt64Key(getKeyOfPK)

	// 缓存中，索引键找不到主键需要查库（此时做索引查行记录）
	if err := cc.cache.TakeEx(&id, indexKey, func(newVal interface{}, expires time.Duration) (err error) {
		id, err = indexQuery(cc.conn, dest)
		if err != nil {
			return
		}

		found = true
		return cc.cache.SetEx(getKeyOfPK(id), dest, expires+safeGapBetweenIndexAndPrimary)
	}); err != nil {
		return err
	}

	// 通过索引已经查到行记录，无须再用主键查询
	if found {
		return nil
	}

	// 通过索引建能直接查到主键，则直接做主键查询
	return cc.cache.Take(dest, getKeyOfPK(id), func(interface{}) error {
		return primaryQuery(cc.conn, dest, id)
	})
}

// 将主键转换为 int64
//
// 解决主键被表示为科学计数法（如2e6），导致缓存无法匹配的问题
func toInt64Key(fn GetKeyOfPKFn) GetKeyOfPKFn {
	return func(primaryKey interface{}) string {
		switch v := primaryKey.(type) {
		case float32:
			return fn(int64(v))
		case float64:
			return fn(int64(v))
		default:
			return fn(primaryKey)
		}
	}
}
