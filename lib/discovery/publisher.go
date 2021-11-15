package discovery

import (
	"git.zc0901.com/go/god/lib/discovery/internal"
	"git.zc0901.com/go/god/lib/lang"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/proc"
	"git.zc0901.com/go/god/lib/syncx"
	"git.zc0901.com/go/god/lib/threading"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type (
	// PublisherOption Publisher 的自定义方法。
	PublisherOption func(client *Publisher)

	// Publisher 用于在指定 key 的 etcd 集群上发布值的发布器。
	Publisher struct {
		endpoints  []string
		key        string
		fullKey    string
		id         int64
		value      string
		leaseID    clientv3.LeaseID
		quit       *syncx.DoneChan
		pauseChan  chan lang.PlaceholderType
		resumeChan chan lang.PlaceholderType
	}
)

// NewPublisher 发挥一个新的 Publisher。
// endpoints 是 etcd 集群的主机。
// key:value 是待发布的键值对。
// opts 是用于自定义 Publisher 的方法。
func NewPublisher(endpoints []string, key, value string, opts ...PublisherOption) *Publisher {
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

func (p *Publisher) KeepAlive() error {
	cli, err := internal.GetRegistry().GetConn(p.endpoints)
	if err != nil {
		return err
	}

	p.leaseID, err = p.register(cli)
	if err != nil {
		return err
	}

	proc.AddWrapUpListener(func() {
		p.Stop()
	})

	return p.keepAliveAsync(cli)
}

func (p *Publisher) Pause() {
	p.pauseChan <- lang.Placeholder
}

func (p *Publisher) Resume() {
	p.resumeChan <- lang.Placeholder
}

func (p *Publisher) Stop() {
	p.quit.Close()
}

func (p *Publisher) register(cli internal.EtcdClient) (clientv3.LeaseID, error) {
	resp, err := cli.Grant(cli.Ctx(), TimeToLive)
	if err != nil {
		return clientv3.NoLease, err
	}

	leaseID := resp.ID
	if p.id > 0 {
		p.fullKey = makeEtcdKey(p.key, p.id)
	} else {
		p.fullKey = makeEtcdKey(p.key, int64(leaseID))
	}
	_, err = cli.Put(cli.Ctx(), p.fullKey, p.value, clientv3.WithLease(leaseID))

	return leaseID, err
}

func (p *Publisher) keepAliveAsync(cli internal.EtcdClient) error {
	ch, err := cli.KeepAlive(cli.Ctx(), p.leaseID)
	if err != nil {
		return err
	}

	threading.GoSafe(func() {
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					p.revoke(cli)
					if err := p.KeepAlive(); err != nil {
						logx.Errorf("KeepAlive: %s", err.Error())
					}
					return
				}
			case <-p.pauseChan:
				logx.Infof("暂停etcd续约, key: %s, value: %s", p.key, p.value)
				p.revoke(cli)
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
				p.revoke(cli)
				return
			}
		}
	})

	return nil
}

func (p *Publisher) revoke(cli internal.EtcdClient) {
	if _, err := cli.Revoke(cli.Ctx(), p.leaseID); err != nil {
		logx.Error(err)
	}
}

// WithId 设定 Publisher 的 id。
func WithId(id int64) PublisherOption {
	return func(publisher *Publisher) {
		publisher.id = id
	}
}

// WithEtcdAccount 自定义 Etcd 账号。
func WithEtcdAccount(user, pass string) PublisherOption {
	return func(client *Publisher) {
		internal.AddAccount(client.endpoints, user, pass)
	}
}
