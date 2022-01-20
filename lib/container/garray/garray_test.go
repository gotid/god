package garray_test

import (
	"fmt"
	"testing"

	"github.com/gotid/god/lib/container/garray"
)

func TestArray_Contains(t *testing.T) {
	ids := []int{1, 2, 3}
	var id int
	id = 1
	fmt.Println(garray.NewIntArrayFrom(ids).Contains(id))
}
