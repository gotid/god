//go:generate mockgen -package internal -destination updatelistener_mock.go -source updatelistener.go UpdateListener

package internal

type (
	// KV 是用于存取给定键值对的 etcd 条目。
	KV struct {
		Key string
		Val string
	}

	// UpdateListener 接口包装 KV 增删方法。
	UpdateListener interface {
		OnAdd(kv KV)
		OnDelete(kv KV)
	}
)
