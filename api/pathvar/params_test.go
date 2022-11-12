package pathvar

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestVars(t *testing.T) {
	expect := map[string]string{
		"a": "1",
		"b": "2",
	}

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	r = WithVars(r, expect)
	actual := Vars(r)
	assert.EqualValues(t, expect, actual)
}

func TestVarsNil(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	assert.Nil(t, Vars(r))
}

func TestContextKey(t *testing.T) {
	ck := contextKey("hellox")
	assert.True(t, strings.Contains(ck.String(), "hellox"))
}
