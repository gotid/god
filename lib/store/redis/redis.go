package redis

import (
	"time"

	red "github.com/go-redis/redis"
	"github.com/gotid/god/lib/breaker"
)

const (
	ClusterMode    = "cluster"
	StandaloneMode = "standalone"
	Nil            = red.Nil

	slowThreshold = 100 * time.Millisecond
)

type (
	Redis struct {
		Addr     string
		Mode     string
		Password string
		tls      bool
		brk      breaker.Breaker
	}

	Client interface {
		red.Cmdable
	}

	// GeoLocation is used with GeoAdd to add geospatial location.
	GeoLocation = red.GeoLocation
	// GeoRadiusQuery is used with GeoRadius to query geospatial index.
	GeoRadiusQuery = red.GeoRadiusQuery
	GeoPos         = red.GeoPos

	Pipeliner = red.Pipeliner

	Pair struct {
		Key   string
		Score float64
	}

	// Option 自定义Redis的方法
	Option func(r *Redis)
)

// New 返回一个根据自定义选项创建的Redis
func New(addr string, opts ...Option) *Redis {
	r := &Redis{
		Addr: addr,
		Mode: StandaloneMode,
		brk:  breaker.NewBreaker(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func NewRedis(addr, mode string, password ...string) *Redis {
	var opts []Option
	if mode == ClusterMode {
		opts = append(opts, Cluster())
	}

	// 为了支持不提供 password 的情况
	for _, v := range password {
		opts = append(opts, WithPassword(v))
	}

	return New(addr, opts...)
}

// Cluster 自定义Redis为集群模式。
func Cluster() Option {
	return func(r *Redis) {
		r.Mode = ClusterMode
	}
}

// WithPassword 自定义Redis密码。
func WithPassword(password string) Option {
	return func(r *Redis) {
		r.Password = password
	}
}

// WithTLS 启用redis的TLS。
func WithTLS() Option {
	return func(r *Redis) {
		r.tls = true
	}
}
