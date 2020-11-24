package token

import (
	"git.zc0901.com/go/god/lib/timex"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const claimHistoryResetDuration = time.Hour * 24

type (
	TokenParser struct {
		resetTime     time.Duration
		resetDuration time.Duration
		history       sync.Map
	}

	ParseOption func(parser *TokenParser)
)

func NewTokenParser(opts ...ParseOption) *TokenParser {
	parser := &TokenParser{
		resetTime:     timex.Now(),
		resetDuration: claimHistoryResetDuration,
	}

	for _, opt := range opts {
		opt(parser)
	}

	return parser
}

func (tp *TokenParser) ParseToken(r *http.Request, secret, prevSecret string) (*jwt.Token, error) {
	var token *jwt.Token
	var err error

	if len(prevSecret) > 0 {
		count := tp.loadCount(secret)
		prevCount := tp.loadCount(prevSecret)

		var first, second string
		if count > prevCount {
			first = secret
			second = prevSecret
		} else {
			first = prevSecret
			second = secret
		}

		token, err = tp.doParseToken(r, first)
		if err != nil {
			token, err = tp.doParseToken(r, second)
			if err != nil {
				return nil, err
			} else {
				tp.incrementCount(second)
			}
		} else {
			tp.incrementCount(first)
		}
	} else {
		token, err = tp.doParseToken(r, secret)
		if err != nil {
			return nil, err
		}
	}

	return token, nil
}

func (tp *TokenParser) doParseToken(r *http.Request, secret string) (*jwt.Token, error) {
	return request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, request.WithParser(newParser()))
}

func (tp *TokenParser) incrementCount(secret string) {
	now := timex.Now()
	if tp.resetTime+tp.resetDuration < now {
		tp.history.Range(func(key, value interface{}) bool {
			tp.history.Delete(key)
			return true
		})
	}

	value, ok := tp.history.Load(secret)
	if ok {
		atomic.AddUint64(value.(*uint64), 1)
	} else {
		var count uint64 = 1
		tp.history.Store(secret, &count)
	}
}

func (tp *TokenParser) loadCount(secret string) uint64 {
	value, ok := tp.history.Load(secret)
	if ok {
		return *value.(*uint64)
	}

	return 0
}

func WithResetDuration(duration time.Duration) ParseOption {
	return func(parser *TokenParser) {
		parser.resetDuration = duration
	}
}

func newParser() *jwt.Parser {
	return &jwt.Parser{
		UseJSONNumber: true,
	}
}
