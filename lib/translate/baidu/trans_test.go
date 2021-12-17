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
