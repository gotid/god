package discovery

import (
	"testing"

	"git.zc0901.com/go/god/lib/discovery/internal"
	"git.zc0901.com/go/god/lib/stringx"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAccount(t *testing.T) {
	endpoints := []string{
		"192.168.0.1:2379",
		"192.168.0.2:2379",
		"192.168.0.3:2379",
	}
	user := "foo" + stringx.Rand()
	RegisterAccount(endpoints, user, "bar")
	account, ok := internal.GetAccount(endpoints)
	assert.True(t, ok)
	assert.Equal(t, user, account.User)
	assert.Equal(t, "bar", account.Pass)
}
