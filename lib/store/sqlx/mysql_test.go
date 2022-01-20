package sqlx

import (
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/gotid/god/lib/breaker"
	"github.com/stretchr/testify/assert"
)

func TestBreakerOnDuplicateEntry(t *testing.T) {
	err := tryOnDuplicateEntryError(t, mysqlAcceptable)
	assert.Equal(t, ErrDuplicateEntryCode, err.(*mysql.MySQLError).Number)
}

func tryOnDuplicateEntryError(t *testing.T, acceptable func(reqError error) bool) error {
	c := &conn{
		brk:    breaker.NewBreaker(),
		accept: acceptable,
	}
	for i := 0; i < 100; i++ {
		assert.NotNil(t, c.brk.DoWithAcceptable(func() error {
			return &mysql.MySQLError{Number: ErrDuplicateEntryCode}
		}, c.acceptable))
	}
	return c.brk.DoWithAcceptable(func() error {
		return &mysql.MySQLError{Number: ErrDuplicateEntryCode}
	}, c.acceptable)
}
