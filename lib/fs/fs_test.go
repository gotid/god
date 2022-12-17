package fs

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os/user"
	"testing"
)

func TestUserHome(t *testing.T) {
	usr, err := user.Current()
	assert.Nil(t, err)
	fmt.Printf("%#v", usr)
}
