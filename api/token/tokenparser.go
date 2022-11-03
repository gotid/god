package token

import (
	"github.com/golang-jwt/jwt/v4/request"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gotid/god/lib/timex"
)

const claimHistoryResetDuration = 24 * time.Hour

type (
	// Parser 是一个用于解析令牌的解析器。
	Parser struct {
		resetTime     time.Duration
		resetDuration time.Duration
		history       sync.Map
	}

	// ParserOption 自定义 Parser。
	ParserOption func(parser *Parser)
)

// NewParser 返回一个 Parser。
func NewParser(opts ...ParserOption) *Parser {
	parser := &Parser{
		resetTime:     timex.Now(),
		resetDuration: claimHistoryResetDuration,
	}

	for _, opt := range opts {
		opt(parser)
	}

	return parser
}

// WithResetDuration 自定义解析器的重置时长。
func WithResetDuration(duration time.Duration) ParserOption {
	return func(parser *Parser) {
		parser.resetDuration = duration
	}
}

// ParseToken 使用传入的 secret、prevSecret 解析给定请求 r 中的令牌。
func (p *Parser) ParseToken(r *http.Request, secret, prevSecret string) (*jwt.Token, error) {
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
			}

			p.incrCount(second)
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

func (p *Parser) loadCount(secret string) uint64 {
	value, ok := p.history.Load(secret)
	if ok {
		return *value.(*uint64)
	}

	return 0
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
		p.history.Range(func(key, value any) bool {
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

func newParser() *jwt.Parser {
	return jwt.NewParser(jwt.WithJSONNumber())
}
