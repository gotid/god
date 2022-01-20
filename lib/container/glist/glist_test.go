package glist_test

import (
	"container/list"
	"fmt"
	"testing"

	"github.com/gotid/god/lib/container/garray"
	"github.com/gotid/god/lib/container/glist"
)

func TestNew(t *testing.T) {
	l := glist.New()
	n := 10
	for i := 0; i < n; i++ {
		l.PushBack(i)
	}
	fmt.Println(l.Len())
	fmt.Println(l.FrontAll())
	fmt.Println(l.BackAll())
	for i := 0; i < n; i++ {
		fmt.Print(l.PopFront())
	}
	l.Clear()
	fmt.Println()
	fmt.Println(l.Len())
}

func TestList_RLockFunc(t *testing.T) {
	// 并发安全列表
	l := glist.NewFrom(garray.NewArrayRange(1, 10, 1).Slice(), true)

	// 从头读
	l.RLockFunc(func(list *list.List) {
		length := list.Len()
		if length > 0 {
			for i, e := 0, list.Front(); i < length; i, e = i+1, e.Next() {
				fmt.Print(e.Value)
			}
		}
	})
	fmt.Println()

	// 从尾读
	l.RLockFunc(func(list *list.List) {
		length := list.Len()
		if length > 0 {
			for i, e := 0, list.Back(); i < length; i, e = i+1, e.Prev() {
				fmt.Print(e.Value)
			}
		}
	})
	fmt.Println()
}

func TestList_IteratorAsc(t *testing.T) {
	l := glist.NewFrom([]interface{}{10, 1, 3, 7, 2}, true)
	l.IteratorDesc(func(e *glist.Element) bool {
		fmt.Print(e.Value)
		return true
	})
}
