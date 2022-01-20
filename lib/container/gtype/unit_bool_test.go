package gtype_test

import (
	"encoding/json"
	"testing"

	"github.com/gotid/god/lib/container/gtype"
	"github.com/gotid/god/lib/gconv"
	"github.com/stretchr/testify/assert"
)

func TestBool_Cas(t *testing.T) {
	v := gtype.NewBool(true)
	vClone := v.Clone()
	assert.Equal(t, true, vClone.Set(false))

	v2 := gtype.NewBool(false)
	v2Clone := v2.Clone()
	assert.Equal(t, false, v2Clone.Set(true))
	assert.Equal(t, true, v2Clone.Val())

	v3 := gtype.NewBool()
	assert.Equal(t, false, v3.Val())
}

func TestBool_MarshalJSON(t *testing.T) {
	v := gtype.NewBool(true)
	b1, err1 := json.Marshal(v)
	b2, err2 := json.Marshal(v.Val())
	assert.Equal(t, nil, err1)
	assert.Equal(t, nil, err2)
	assert.Equal(t, b1, b2)
}

func TestBool_UnmarshalJSON(t *testing.T) {
	v := gtype.NewBool()
	err := json.Unmarshal([]byte("true"), &v)
	assert.Nil(t, err)
	assert.Equal(t, true, v.Val())

	err = json.Unmarshal([]byte("false"), &v)
	assert.Nil(t, err)
	assert.Equal(t, false, v.Val())

	err = json.Unmarshal([]byte("1"), &v)
	assert.Nil(t, err)
	assert.Equal(t, true, v.Val())

	err = json.Unmarshal([]byte("0"), &v)
	assert.Nil(t, err)
	assert.Equal(t, false, v.Val())
}

func TestBool_UnmarshalValue(t *testing.T) {
	type V struct {
		Name string
		Var  *gtype.Bool
	}

	var v *V
	err := gconv.Struct(map[string]interface{}{
		"name": "john",
		"var":  "true",
	}, &v)

	assert.Nil(t, err)
	assert.Equal(t, v.Name, "john")
	assert.Equal(t, v.Var.Val(), true)
}
