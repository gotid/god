package gvalid_test

import (
	"testing"

	"github.com/gotid/god/lib/g"
	"github.com/gotid/god/lib/gvalid"
	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	val1 := 0
	rule := "aaa:6,16"
	err1 := gvalid.Check(val1, rule, nil)
	assert.Equal(t, "invalid rules: aaa:6,16", err1.Error())
}

func TestRequired(t *testing.T) {
	e := gvalid.Check("1", "required", nil)
	assert.Nil(t, e)

	e = gvalid.Check("哈哈", "required-if: id,1,age,18", nil, g.Map{"id": 1, "age": 18})
	assert.NotNil(t, e)
}
