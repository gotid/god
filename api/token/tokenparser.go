package token

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"git.zc0901.com/go/god/lib/timex"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

const claimHistoryResetDuration = 24 * time.Hour

type (
	// Parser 是一个令牌解析器。
	Parser struct {
		resetTime     time.Duration
		resetDuration time.Duration
		history       sync.Map
	}

	// ParseOption 是一个自定义 Parser 的函数。
	ParseOption func(parser *Parser)
)

// NewTokenParser 返回一个令牌解析器。
func NewTokenParser(opts ...ParseOption) *Parser {
	parser := &Parser{
		resetTime:     timex.Now(),
		resetDuration: claimHistoryResetDuration,
	}

	for _, opt := range opts {
		opt(parser)
	}

	return parser
}

// Parse 从请求中使用密钥和上一个密钥来解析令牌。
func (p *Parser) Parse(r *http.Request, secret, prevSecret string) (*jwt.Token, error) {
	var token *jwt.Token
	var err error

	if len(prevSecret) > 0 {
		count := p.loadCount(secret)
		prevCount := p.loadCount(prevSecret)

		var first, second string
		if count > prevCount {
			first = secret
			second = prevSecret
		} else {
			first = prevSecret
			second = secret
		}

		token, err = p.doParseToken(r, first)
		if err != nil {
			token, err = p.doParseToken(r, second)
			if err != nil {
				return nil, err
			} else {
				p.incrCount(second)
			}
		} else {
			p.incrCount(first)
		}
	} else {
		token, err = p.doParseToken(r, secret)
		if err != nil {
			return nil, err
		}
	}

	return token, nil
}

func (p *Parser) doParseToken(r *http.Request, secret string) (*jwt.Token, error) {
	return request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, request.WithParser(newParser()))
}

func (p *Parser) incrCount(secret string) {
	now := timex.Now()
	if p.resetTime+p.resetDuration < now {
		p.history.Range(func(key, value interface{}) bool {
			p.history.Delete(key)
			return true
		})
	}

	value, ok := p.history.Load(secret)
	if ok {
		atomic.AddUint64(value.(*uint64), 1)
	} else {
		var count uint64 = 1
		p.history.Store(secret, &count)
	}
}

func (p *Parser) loadCount(secret string) uint64 {
	value, ok := p.history.Load(secret)
	if ok {
		return *value.(*uint64)
	}

	return 0
}

// WithResetDuration 自定义令牌重置时长。
func WithResetDuration(duration time.Duration) ParseOption {
	return func(parser *Parser) {
		parser.resetDuration = duration
	}
}

func newParser() *jwt.Parser {
	return &jwt.Parser{
		UseJSONNumber: true,
	}
}
