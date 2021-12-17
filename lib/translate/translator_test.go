package translate

import (
	"fmt"
	"testing"
)

func TestBaidu(t *testing.T) {
	Baidu.Time(true)
	fmt.Println(Baidu.Zh2En("你好"))
	fmt.Println(Baidu.En2Zh("European living room"))
	fmt.Println(Baidu.ToZh("Holism Retreat by Studio Tate"))
	fmt.Println(Baidu.ToZh(`The text sample should be bigger then 200kb
and can be "dirty" (special chars, lists, etc.),
but the language should not change for long parts.`))
}

func BenchmarkBaidu(b *testing.B) {
	for i := 0; i < 100; i++ {
		fmt.Println(Baidu.Zh2En("你好"))
	}
}
