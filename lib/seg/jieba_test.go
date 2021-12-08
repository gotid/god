package seg

import (
	"fmt"
	"strings"

	"github.com/yanyiwu/gojieba"
)

import "testing"

func TestJieba(t *testing.T) {
	var s string
	var words []string

	x := gojieba.NewJieba(gojieba.DICT_PATH, gojieba.HMM_PATH, "dict.txt")
	defer x.Free()

	s = "春天的花开秋天的风以及冬天的落阳"
	s = "红色布艺条纹沙发 白色窗帘 设计或灵感"
	// s = "中国市民有十三亿人口"
	// s = "床头 柜子"
	s = "红色沙发 白色窗帘"
	words = x.CutAll(s)
	fmt.Println("全模式", strings.Join(words, "/"))

	words = x.Cut(s, true)
	fmt.Println("精确模式", strings.Join(words, "/"))

	words = x.CutForSearch(s, true)
	fmt.Println("搜索模式", strings.Join(words, "/"))
	println()

	words = x.Tag(s)
	fmt.Println("词性", words)

	weights := x.ExtractWithWeight(s, 10)
	fmt.Println("权重", weights)

	wordInfos := x.Tokenize(s, gojieba.DefaultMode, true)
	fmt.Println("位置", wordInfos)
}
