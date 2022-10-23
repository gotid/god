package discov

import (
	"github.com/gotid/god/lib/discov/internal"
	"github.com/gotid/god/lib/lang"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/proc"
	"github.com/gotid/god/lib/syncx"
	"github.com/gotid/god/lib/threading"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type (
	// PubOption 自定义 Publisher 的方法。
	PubOption func(p *Publisher)

	// Publisher 基于给定键向 etcd 集群发送值。
	Publisher struct {
		endpoints  []string
		key        string
		fullKey    string
		id         int64
		value      string
		lease      clientv3.LeaseID
		quit       *syncx.DoneChan
		pauseChan  chan lang.PlaceholderType
		resumeChan chan lang.PlaceholderType
	}
)

// NewPublisher 返回一个发布器 Publisher。
// endpoints 是 etcd 集群的主机。
// key:value 是待发布的键值对。
// opts 是自定义发布器的方法。
func NewPublisher(endpoints []string, key, value string, opts ...PubOption) *Publisher {
	publisher := &Publisher{
		endpoints:  endpoints,
		key:        key,
		value:      value,
		quit:       syncx.NewDoneChan(),
		pauseChan:  make(chan lang.PlaceholderType),
		resumeChan: make(chan lang.PlaceholderType),
	}

	for _, opt := range opts {
		opt(publisher)
	}

	return publisher
}

// KeepAlive 保持键值对为存活状态。
func (p *Publisher) KeepAlive() error {
	client, err := internal.GetRegistry().GetConn(p.endpoints)
	if err != nil {
		return err
	}

	p.lease, err = p.register(client)
	if err != nil {
		return err
	}
	proc.AddWrapUpListener(func() {
		p.Stop()
	})

	return p.keepAliveAsync(client)
}

// Pause 暂停更新
func (p *Publisher) Pause() {
	p.pauseChan <- lang.Placeholder
}

// Resume 继续更新
func (p *Publisher) Resume() {
	p.resumeChan <- lang.Placeholder
}

// Stop 停止续订并取消注册。
func (p *Publisher) Stop() {
	p.quit.Close()
}

func (p *Publisher) register(client internal.EtcdClient) (clientv3.LeaseID, error) {
	resp, err := client.Grant(client.Ctx(), TimeToLive)
	if err != nil {
		return clientv3.NoLease, err
	}

	lease := resp.ID
	if p.id > 0 {
		p.fullKey = makeEtcdKey(p.key, p.id)
	} else {
		p.fullKey = makeEtcdKey(p.key, int64(lease))
	}
	_, err = client.Put(client.Ctx(), p.fullKey, p.value, clientv3.WithLease(lease))

	return lease, err
}

func (p *Publisher) keepAliveAsync(client internal.EtcdClient) error {
	ch, err := client.KeepAlive(client.Ctx(), p.lease)
	if err != nil {
		return err
	}

	threading.GoSafe(func() {
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					p.revoke(client)
					if err := p.KeepAlive(); err != nil {
						logx.Errorf("KeepAlive: %s", err.Error())
					}
					return
				}
			case <-p.pauseChan:
				logx.Infof("已暂停 etcd 续订，key: %s, value: %s", p.key, p.value)
				p.revoke(client)
				select {
				case <-p.resumeChan:
					if err := p.KeepAlive(); err != nil {
						logx.Errorf("KeepAlive: %s", err.Error())
					}
					return
				case <-p.quit.Done():
					return
				}
			case <-p.quit.Done():
				p.revoke(client)
				return
			}
		}
	})

	return nil
}

func (p *Publisher) revoke(client internal.EtcdClient) {
	if _, err := client.Revoke(client.Ctx(), p.lease); err != nil {
		logx.Error(err)
	}
}

// WithId 自定义发布器 Publisher 的 id。
func WithId(id int64) PubOption {
	return func(p *Publisher) {
		p.id = id
	}
}

// WithPubEtcdAccount 自定义 etcd 的用户名/密码。
func WithPubEtcdAccount(user, pass string) PubOption {
	return func(p *Publisher) {
		RegisterAccount(p.endpoints, user, pass)
	}
}

// WithPubEtcdTLS 自定义 etcd 的 TLS 证书。
func WithPubEtcdTLS(certFile, certKeyFile, caFile string, insecureSkipVerify bool) PubOption {
	return func(p *Publisher) {
		logx.Must(RegisterTLS(p.endpoints, certFile, certKeyFile, caFile, insecureSkipVerify))
	}
}
