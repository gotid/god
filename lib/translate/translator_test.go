package translate

import (
	"fmt"
	"testing"
)

func TestBaidu(t *testing.T) {
	// fmt.Println(Baidu.Zh2En("你好"))
	fmt.Println(Baidu.En2Zh("European living room"))
}

func BenchmarkBaidu(b *testing.B) {
	for i := 0; i < 100; i++ {
		fmt.Println(Baidu.Zh2En("你好"))
	}
}
