package search

import (
	"errors"
	"fmt"
)

const (
	slash = '/'
	colon = ':'
)

var (
	// errNotFromRoot 意为路径没有以 slash 开头。
	errNotFromRoot = errors.New("路径必须以 / 开始")
	// errEmptyItem 意为添加了一个空白的路由条目。
	errEmptyItem = errors.New("路由的条目不能为空")
	// errDuplicateItem 意为添加了一个重复的路由条目。
	errDuplicateItem = errors.New("路由条目不能重复")
	// errDuplicateSlash 意为添加了一个多个斜线开头的路由条目。
	errDuplicateSlash = errors.New("路由开头的斜线不能超过1个")
	// errInvalidState 意为搜索树处于一个无效的状态。
	errInvalidState = errors.New("搜索树处于一个无效的状态")

	// NotFound 用于保存未找到的结果。
	NotFound Result
)

type (
	// Tree 为一颗用于搜索的树。
	Tree struct {
		root *node
	}

	// Result 从树中搜索到的结果。
	Result struct {
		Item   interface{}
		Params map[string]string
	}

	node struct {
		item     interface{}
		children [2]map[string]*node
	}

	innerResult struct {
		key   string
		value string
		named bool
		found bool
	}
)

// NewTree 返回一个 Tree 的示例。
func NewTree() *Tree {
	return &Tree{
		root: newNode(nil),
	}
}

// Add 添加一个路由关联的条目到树中。
func (t *Tree) Add(route string, item interface{}) error {
	if len(route) == 0 || route[0] != slash {
		return errNotFromRoot
	}

	if item == nil {
		return errEmptyItem
	}

	err := add(t.root, route[1:], item)
	switch err {
	case errDuplicateItem:
		return duplicatedItem(route)
	case errDuplicateSlash:
		return duplicatedSlash(route)
	default:
		return err
	}
}

// Search 搜索给定路由关联的条目。
func (t *Tree) Search(route string) (Result, bool) {
	if len(route) == 0 || route[0] != slash {
		return NotFound, false
	}

	var result Result
	ok := t.next(t.root, route[1:], &result)

	return result, ok
}

func (t *Tree) next(n *node, route string, result *Result) bool {
	if len(route) == 0 && n.item != nil {
		result.Item = n.item
		return true
	}

	for i := range route {
		if route[i] != slash {
			continue
		}

		token := route[:i]
		return n.forEach(func(k string, v *node) bool {
			r := match(k, token)
			if !r.found || !t.next(v, route[i+1:], result) {
				return false
			}
			if r.named {
				addParam(result, r.key, r.value)
			}

			return true
		})
	}

	return n.forEach(func(k string, v *node) bool {
		if r := match(k, route); r.found && v.item != nil {
			result.Item = v.item
			if r.named {
				addParam(result, r.key, r.value)
			}
			return true
		}

		return false
	})
}

func addParam(result *Result, key, value string) {
	if result.Params == nil {
		result.Params = make(map[string]string)
	}

	result.Params[key] = value
}

func match(pat, token string) innerResult {
	if pat[0] == colon {
		return innerResult{
			key:   pat[1:],
			value: token,
			named: true,
			found: true,
		}
	}

	return innerResult{
		found: pat == token,
	}
}

func add(n *node, route string, item interface{}) error {
	if len(route) == 0 {
		if n.item != nil {
			return errDuplicateItem
		}

		n.item = item
		return nil
	}

	if route[0] == slash {
		return errDuplicateSlash
	}

	for i := range route {
		if route[i] != slash {
			continue
		}

		token := route[:i]
		children := n.getChildren(token)
		if child, ok := children[token]; ok {
			if child != nil {
				return add(child, route[i+1:], item)
			}

			return errInvalidState
		}

		child := newNode(nil)
		children[token] = child
		return add(child, route[i+1:], item)
	}

	children := n.getChildren(route)
	if child, ok := children[route]; ok {
		if child.item != nil {
			return errDuplicateItem
		}

		child.item = item
	} else {
		children[route] = newNode(item)
	}

	return nil
}

func (n *node) getChildren(route string) map[string]*node {
	if len(route) > 0 && route[0] == colon {
		return n.children[1]
	}

	return n.children[0]
}

func (n *node) forEach(fn func(string, *node) bool) bool {
	for _, child := range n.children {
		for k, v := range child {
			if fn(k, v) {
				return true
			}
		}
	}

	return false
}

func newNode(item interface{}) *node {
	return &node{
		item: item,
		children: [2]map[string]*node{
			make(map[string]*node),
			make(map[string]*node),
		},
	}
}

func duplicatedSlash(route string) error {
	return fmt.Errorf("重复的斜线 %s", route)
}

func duplicatedItem(route string) error {
	return fmt.Errorf("重复的路由条目 %s", route)
}
