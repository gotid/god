package baidu

import (
	"fmt"
	"testing"
)

func TestTranslate_Translate(t *testing.T) {
	trans := Must()

	result := trans.Zh2En("欧式客厅")
	fmt.Println(result)

	result = trans.En2Zh("European living room")
	fmt.Println(result)
}

func TestTranslate_Detect(t *testing.T) {
	trans := Must()

	fmt.Println(trans.ToEn("我爱你"))

	fmt.Println(trans.ToEn("爱してる"))

	fmt.Println(trans.ToZh("사 랑 해 요"))

	fmt.Println(trans.ToZh("I love you"))

	fmt.Println(trans.ToZh("Je t'aime"))

	fmt.Println(trans.ToZh("te amo,tequiero"))

	fmt.Println(trans.ToZh("te amo,tequiero"))

	fmt.Println(trans.ToZh("TI AMO"))

	fmt.Println(trans.ToZh("Я люблю тебя"))

	fmt.Println(trans.ToZh("Я тебе кохаю! "))
}
