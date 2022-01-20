package weapp

import (
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	app := New(nil, "")
	fmt.Println(app)
}
