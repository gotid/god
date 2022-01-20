package codec

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	var b bytes.Buffer
	for i := 0; i < 10000; i++ {
		fmt.Fprint(&b, i)
	}

	bs := Gzip(b.Bytes())
	actual, err := Gunzip(bs)

	assert.Nil(t, err)
	assert.True(t, len(bs) < b.Len())
	assert.Equal(t, b.Bytes(), actual)

	fmt.Println(b.Len())
	fmt.Println(len(bs))
}
