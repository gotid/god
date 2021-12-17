package translate

import (
	"fmt"
	"testing"
)

func TestBaidu(t *testing.T) {
	fmt.Println(Baidu.Zh2En("你好"))
	fmt.Println(Baidu.En2Zh("European living room"))
}
