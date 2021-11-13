package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"git.zc0901.com/go/god/lib/stringx"
)

func TestAccount(t *testing.T) {
	endpoints := []string{
		"192.168.0.1:2379",
		"192.168.0.2:2379",
		"192.168.0.3:2379",
	}
	user := "foo" + stringx.Rand()
	pass := "bar" + stringx.Rand()
	anotherPass := "any"

	_, ok := GetAccount(endpoints)
	assert.False(t, ok)

	AddAccount(endpoints, user, pass)
	account, ok := GetAccount(endpoints)
	assert.True(t, ok)
	assert.Equal(t, user, account.User)
	assert.Equal(t, pass, account.Pass)

	AddAccount(endpoints, user, anotherPass)
	account, ok = GetAccount(endpoints)
	assert.True(t, ok)
	assert.Equal(t, user, account.User)
	assert.Equal(t, anotherPass, account.Pass)
}
