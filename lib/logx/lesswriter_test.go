package logx

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLessWriter(t *testing.T) {
	var b strings.Builder
	w := newLessWriter(&b, 1000)
	for i := 0; i < 10; i++ {
		_, err := w.Write([]byte("hello" + strconv.Itoa(i)))
		assert.Nil(t, err)
	}

	// assert.Equal(t, "hello", b.String())
	fmt.Println(b.String())
}
