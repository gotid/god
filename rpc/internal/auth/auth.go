package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/lib/store/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const defaultExpiration = time.Minute * 5

type Authenticator struct {
	store  *redis.Redis      // redis存储（二级缓存）
	key    string            // redis 存储鉴权信息的hash key
	cache  *collection.Cache // 进程内缓存（一级缓存）
	strict bool              // 严格模式将返回框架内部错误信息
}

func NewAuthenticator(store *redis.Redis, key string, strict bool) (*Authenticator, error) {
	cache, err := collection.NewCache(defaultExpiration)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		store:  store,
		key:    key,
		cache:  cache,
		strict: strict,
	}, nil
}

// Authenticate 通过metadata中传递的app/token到进程或redis中进行对比鉴权
func (a *Authenticator) Authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, missingMetadata)
	}

	apps, tokens := md[appKey], md[tokenKey]
	if len(apps) == 0 || len(tokens) == 0 {
		return status.Errorf(codes.Unauthenticated, missingMetadata)
	}

	app, token := apps[0], tokens[0]
	if len(app) == 0 || len(token) == 0 {
		return status.Error(codes.Unauthenticated, missingMetadata)
	}

	return a.validate(app, token)
}

// validate 验证app/token有效性
func (a *Authenticator) validate(app string, token string) error {
	// 先从进程级内存查找app对应的token，找不到则去redis中查找
	except, err := a.cache.Take(app, func() (interface{}, error) {
		/*
			{
				"key1": {"app1": "xxx", "app2": "yyy"},
				"key2": {"app3": "xxx", "app4": "yyy"},
			}
		*/
		return a.store.HGet(a.key, app)
	})
	if err != nil {
		if a.strict {
			return status.Error(codes.Internal, fmt.Sprintf("[RPC鉴权] %s", err.Error()))
		} else {
			return nil
		}
	}

	if token != except {
		return status.Error(codes.Unauthenticated, accessDenied)
	}

	return nil
}
