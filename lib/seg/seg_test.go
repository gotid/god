package seg

import (
	"fmt"
	"testing"
	"time"
)

var (
	// 删除词
	StopWords = []string{
		"设计",
		"灵感",
		"风格",
	}

	// 合并词
	CombineWords = []string{
		"benjamin moore",
		"setting wall",
	}

	// 合成词映射
	CombineMap = map[string]string{
		"art deco": "art-deco",
		//"复古 未来":    "复古未来主义",
	}

	// 同义词映射
	SynonymMap = map[string]string{
		"玄关":          "门厅",
		"白":           "白色",
		"绿":           "绿色",
		"粉":           "粉色",
		"蓝":           "蓝色",
		"红":           "红色",
		"黄":           "黄色",
		"灰":           "灰色",
		"紫":           "紫色",
		"橙":           "橙色",
		"橘":           "橘色",
		"棕":           "棕色",
		"黑":           "黑色",
		"大地色":         "大地色系",
		"马卡龙":         "马卡龙色",
		"莫兰迪":         "莫兰迪色",
		"荧光":          "荧光色",
		"金":           "金色",
		"银":           "银色",
		"settingwall": "背景墙",
		"诧寂":          "侘寂",
		"复古未来":        "复古未来主义",
	}
)

func TestSeg_CutForSearch(t *testing.T) {
	segmenter := NewSegmenter(
		"dict.txt",
		StopWords,
		CombineWords,
		CombineMap,
		SynonymMap,
		[][]string{},
		false,
	)

	q := "客厅 现代 轻奢 红色房子"
	// q = "沙发 现代 白客厅"
	// q = "沙发区怎么搭配大理石"
	// q = "酒店 大理石 沙发"
	// q = "卫生间 白色瓷砖墙面"
	// q = "卫生间 墙砖"
	// q = "imola白色瓷砖"
	// q = "瓷砖 fasfj 😄 白色"
	// q = "岛台 瓷砖"
	q = "红色沙发 白色窗帘"
	start := time.Now().UnixMicro()
	keywords := segmenter.CutForSearch(q, 6, false)
	fmt.Println("⌚️ 耗时", time.Now().UnixMicro()-start, "微秒")

	for i, keyword := range keywords {
		fmt.Println(i, keyword.Word, keyword.Tag, keyword.Distance, keyword.Weight)
	}
}
