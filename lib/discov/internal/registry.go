package internal

import (
	"context"
	"fmt"
	"github.com/gotid/god/lib/contextx"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	registry = Registry{
		clusters: make(map[string]*cluster),
	}
	connManager = syncx.NewResourceManager()
)

// Registry 是一个管理 etcd 客户端连接的注册表。
type Registry struct {
	clusters map[string]*cluster
	lock     sync.Mutex
}

// GetRegistry 返回一个全局注册表。
func GetRegistry() *Registry {
	return &registry
}

// GetConn 返回一个给定 etcd 端点的可复用客户端连接。
func (r *Registry) GetConn(endpoints []string) (EtcdClient, error) {
	c, _ := r.getCluster(endpoints)
	return c.getClient()
}

// Monitor 监控给定 etcd 端点的 key，通过 UpdateListener 进行通知。
func (r *Registry) Monitor(endpoints []string, key string, l UpdateListener) error {
	c, exists := r.getCluster(endpoints)
	if exists {
		kvs := c.getCurrent(key)
		for _, kv := range kvs {
			l.OnAdd(kv)
		}
	}

	return c.monitor(key, l)
}

func (r *Registry) getCluster(endpoints []string) (c *cluster, exists bool) {
	clusterKey := getClusterKey(endpoints)
	r.lock.Lock()
	defer r.lock.Unlock()
	c, exists = r.clusters[clusterKey]
	if !exists {
		c = newCluster(endpoints)
		r.clusters[clusterKey] = c
	}

	return
}

type cluster struct {
	endpoints  []string
	key        string
	values     map[string]map[string]string
	listeners  map[string][]UpdateListener
	watchGroup *threading.RoutineGroup
	done       chan lang.PlaceholderType
	lock       sync.Mutex
}

func newCluster(endpoints []string) *cluster {
	return &cluster{
		endpoints:  endpoints,
		key:        getClusterKey(endpoints),
		values:     make(map[string]map[string]string),
		listeners:  make(map[string][]UpdateListener),
		watchGroup: threading.NewRoutineGroup(),
		done:       make(chan lang.PlaceholderType),
	}
}

func (c *cluster) context(cli EtcdClient) context.Context {
	return contextx.ValueOnlyFrom(cli.Ctx())
}

func (c *cluster) getClient() (EtcdClient, error) {
	val, err := connManager.Get(c.key, func() (io.Closer, error) {
		return c.newClient()
	})
	if err != nil {
		return nil, err
	}

	return val.(EtcdClient), nil
}

func (c *cluster) newClient() (EtcdClient, error) {
	cli, err := NewClient(c.endpoints)
	if err != nil {
		return nil, err
	}

	go c.watchConnState(cli)

	return cli, nil
}

func (c *cluster) watchConnState(cli EtcdClient) {
	watcher := newStateWatcher()
	watcher.addListener(func() {
		go c.reload(cli)
	})
	watcher.watch(cli.ActiveConnection())
}

func (c *cluster) reload(cli EtcdClient) {
	c.lock.Lock()
	close(c.done)
	c.watchGroup.Wait()
	c.done = make(chan lang.PlaceholderType)
	c.watchGroup = threading.NewRoutineGroup()
	var keys []string
	for k := range c.listeners {
		keys = append(keys, k)
	}
	c.lock.Unlock()

	for _, key := range keys {
		k := key
		c.watchGroup.Run(func() {
			rev := c.load(cli, k)
			c.watch(cli, k, rev)
		})
	}
}

func (c *cluster) load(cli EtcdClient, key string) int64 {
	var resp *clientv3.GetResponse
	for {
		var err error
		ctx, cancel := context.WithTimeout(c.context(cli), RequestTimeout)
		resp, err = cli.Get(ctx, makeKeyPrefix(key), clientv3.WithPrefix())
		cancel()
		if err == nil {
			break
		}

		logx.Error(err)
		time.Sleep(coolDownInterval)
	}

	var kvs []KV
	for _, v := range resp.Kvs {
		kvs = append(kvs, KV{
			Key: string(v.Key),
			Val: string(v.Value),
		})
	}

	c.handleChanges(key, kvs)

	return resp.Header.Revision
}

func (c *cluster) handleChanges(key string, kvs []KV) {
	var add []KV
	var remove []KV

	c.lock.Lock()
	listeners := append([]UpdateListener(nil), c.listeners[key]...)
	vals, ok := c.values[key]
	if !ok {
		add = kvs
		vals = make(map[string]string)
		for _, kv := range kvs {
			vals[kv.Key] = kv.Val
		}
		c.values[key] = vals
	} else {
		m := make(map[string]string)
		for _, kv := range kvs {
			m[kv.Key] = kv.Val
		}
		for k, v := range vals {
			if vals, ok := m[k]; !ok || v != vals {
				remove = append(remove, KV{
					Key: k,
					Val: v,
				})
			}
		}
		for k, v := range m {
			if val, ok := vals[k]; !ok || v != val {
				add = append(add, KV{
					Key: k,
					Val: v,
				})
			}
		}
	}
	c.lock.Unlock()

	// 处理新增
	for _, kv := range add {
		for _, l := range listeners {
			l.OnAdd(kv)
		}
	}

	// 处理移除
	for _, kv := range remove {
		for _, l := range listeners {
			l.OnDelete(kv)
		}
	}
}

func (c *cluster) watch(cli EtcdClient, key string, rev int64) {
	for {
		if c.watchStream(cli, key, rev) {
			return
		}
	}
}

func (c *cluster) watchStream(cli EtcdClient, key string, rev int64) bool {
	var watchCh clientv3.WatchChan
	if rev != 0 {
		watchCh = cli.Watch(
			clientv3.WithRequireLeader(c.context(cli)),
			makeKeyPrefix(key),
			clientv3.WithPrefix(),
			clientv3.WithRev(rev+1),
		)
	} else {
		watchCh = cli.Watch(
			clientv3.WithRequireLeader(c.context(cli)),
			makeKeyPrefix(key),
			clientv3.WithPrefix(),
		)
	}

	for {
		select {
		case resp, ok := <-watchCh:
			if !ok {
				logx.Error("etcd 监控器通道已被关闭")
				return false
			}
			if resp.Canceled {
				logx.Errorf("etcd 监控器通道已被取消，错误：%v", resp.Err())
				return false
			}
			if resp.Err() != nil {
				logx.Errorf("etcd 监控器通道错误：%v", resp.Err())
				return false
			}

			c.handleWatchEvents(key, resp.Events)
		case <-c.done:
			return true
		}
	}
}

func (c *cluster) handleWatchEvents(key string, events []*clientv3.Event) {
	c.lock.Lock()
	listeners := append([]UpdateListener(nil), c.listeners[key]...)
	c.lock.Unlock()

	for _, event := range events {
		switch event.Type {
		case clientv3.EventTypePut:
			c.lock.Lock()
			if vals, ok := c.values[key]; ok {
				vals[string(event.Kv.Key)] = string(event.Kv.Value)
			} else {
				c.values[key] = map[string]string{string(event.Kv.Key): string(event.Kv.Value)}
			}
			c.lock.Unlock()
			for _, l := range listeners {
				l.OnAdd(KV{
					Key: string(event.Kv.Key),
					Val: string(event.Kv.Value),
				})
			}
		case clientv3.EventTypeDelete:
			c.lock.Lock()
			if vals, ok := c.values[key]; ok {
				delete(vals, string(event.Kv.Key))
			}
			c.lock.Unlock()
			for _, l := range listeners {
				l.OnDelete(KV{
					Key: string(event.Kv.Key),
					Val: string(event.Kv.Value),
				})
			}
		default:
			logx.Errorf("未知事件类型：%v", event.Type)
		}
	}
}

func (c *cluster) getCurrent(key string) []KV {
	c.lock.Lock()
	defer c.lock.Unlock()

	var kvs []KV
	for k, v := range c.values[key] {
		kvs = append(kvs, KV{
			Key: k,
			Val: v,
		})
	}

	return kvs
}

func (c *cluster) monitor(key string, l UpdateListener) error {
	c.lock.Lock()
	c.listeners[key] = append(c.listeners[key], l)
	c.lock.Unlock()

	cli, err := c.getClient()
	if err != nil {
		return err
	}

	rev := c.load(cli, key)
	c.watchGroup.Run(func() {
		c.watch(cli, key, rev)
	})

	return nil
}

// NewClient 创建给定端点的 etcd 集群。
func NewClient(endpoints []string) (EtcdClient, error) {
	config := clientv3.Config{
		Endpoints:            endpoints,
		AutoSyncInterval:     autoSyncInterval,
		DialTimeout:          DialTimeout,
		DialKeepAliveTime:    dailyKeepAliveTime,
		DialKeepAliveTimeout: DialTimeout,
		RejectOldCluster:     true,
		PermitWithoutStream:  true,
	}
	if account, ok := GetAccount(endpoints); ok {
		config.Username = account.User
		config.Password = account.Pass
	}
	if tlsConfig, ok := GetTLS(endpoints); ok {
		config.TLS = tlsConfig
	}

	client, err := clientv3.New(config)
	return client, err
}

func makeKeyPrefix(key string) string {
	return fmt.Sprintf("%s%c", key, Delimiter)
}

func getClusterKey(endpoints []string) string {
	sort.Strings(endpoints)
	return strings.Join(endpoints, endpointsSeparator)
}
