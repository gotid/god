package discovery

import (
	"testing"

	"github.com/gotid/god/lib/discovery/internal"
	"github.com/stretchr/testify/assert"
)

const (
	actionAdd = iota
	actionDel
)

func TestContainer(t *testing.T) {
	type action struct {
		act int
		key string
		val string
	}

	tests := []struct {
		name   string
		do     []action
		expect []string
	}{
		{
			name: "添加一个",
			do: []action{
				{
					act: actionAdd,
					key: "第一个",
					val: "a",
				},
			},
			expect: []string{"a"},
		},
		{
			name: "添加两个",
			do: []action{
				{
					act: actionAdd,
					key: "第一个",
					val: "a",
				},
				{
					act: actionAdd,
					key: "第二个",
					val: "b",
				},
			},
			expect: []string{"a", "b"},
		},
		{
			name: "添加两个，删除一个",
			do: []action{
				{
					act: actionAdd,
					key: "第一个",
					val: "a",
				},
				{
					act: actionAdd,
					key: "第二个",
					val: "b",
				},
				{
					act: actionDel,
					key: "第一个",
				},
			},
			expect: []string{"b"},
		},
		{
			name: "添加两个，删除两个",
			do: []action{
				{
					act: actionAdd,
					key: "第一个",
					val: "a",
				},
				{
					act: actionAdd,
					key: "第二个",
					val: "b",
				},
				{
					act: actionDel,
					key: "第一个",
				},
				{
					act: actionDel,
					key: "第二个",
				},
			},
			expect: []string{},
		},
		{
			name: "add three, dup values, delete two",
			do: []action{
				{
					act: actionAdd,
					key: "first",
					val: "a",
				},
				{
					act: actionAdd,
					key: "second",
					val: "b",
				},
				{
					act: actionAdd,
					key: "third",
					val: "a",
				},
				{
					act: actionDel,
					key: "first",
				},
				{
					act: actionDel,
					key: "second",
				},
			},
			expect: []string{"a"},
		},
		{
			name: "add three, dup values, delete two, delete not added",
			do: []action{
				{
					act: actionAdd,
					key: "first",
					val: "a",
				},
				{
					act: actionAdd,
					key: "second",
					val: "b",
				},
				{
					act: actionAdd,
					key: "third",
					val: "a",
				},
				{
					act: actionDel,
					key: "first",
				},
				{
					act: actionDel,
					key: "second",
				},
				{
					act: actionDel,
					key: "forth",
				},
			},
			expect: []string{"a"},
		},
	}

	excludes := []bool{true, false}
	for _, test := range tests {
		for _, exclude := range excludes {
			t.Run(test.name, func(t *testing.T) {
				var changed bool
				c := newContainer(exclude)
				c.addListener(func() {
					changed = true
				})
				assert.Nil(t, c.getValues())
				assert.False(t, changed)

				for _, order := range test.do {
					if order.act == actionAdd {
						c.OnAdd(internal.KV{Key: order.key, Val: order.val})
					} else {
						c.OnDelete(internal.KV{Key: order.key, Val: order.val})
					}
				}

				assert.True(t, changed)
				assert.True(t, c.dirty.True())
				assert.ElementsMatch(t, test.expect, c.getValues())
				assert.False(t, c.dirty.True())
				assert.ElementsMatch(t, test.expect, c.getValues())
			})
		}
	}
}
