package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithRmSpace(t *testing.T) {
	q := " 北欧  客厅 \n 现代       沙发  "
	c := NewCleaner(q)
	assert.Equal(t, "北欧 客厅 现代 沙发", c.RemoveSpace().String())
}

func TestWithRmEmoji(t *testing.T) {
	q := "北欧沙发👦✋👏🏻"
	c := NewCleaner(q)
	s := c.RemoveEmoji().String()
	assert.Equal(t, "北欧沙发", s)
}

func BenchmarkCleaner_RemoveEmoji(b *testing.B) {
	b.ReportAllocs()
	q := "北欧沙发🛋👏"
	c := NewCleaner(q)
	for i := 0; i < b.N; i++ {
		c.RemoveEmoji().String()
	}
}

func TestWithString(t *testing.T) {
	q := "IMOLA 白瓷砖"
	c := NewCleaner(q)
	assert.Equal(t, "imola 白瓷砖", c.String())
}

func TestWithRmStopwords(t *testing.T) {
	q := "北欧设计灵感"
	c := NewCleaner(q)
	assert.Equal(t, "北欧设计灵感", c.StopWords().String())

	c = NewCleaner(q, WithStopWords("设计", "灵感", "风格"))
	assert.Equal(t, "北欧", c.StopWords().String())
}

func TestWithCombineWords(t *testing.T) {
	q := "art deco 复古 未来"
	c := NewCleaner(q, WithCombineWords(
		[]string{"art deco"},
		map[string]string{
			"复古 未来": "复古未来",
		},
	))
	s := c.CombineWords().String()
	assert.Equal(t, "art-deco 复古未来", s)
}

func TestWithSynonymWords(t *testing.T) {
	q := "白玄关"
	c := NewCleaner(q, WithSynonymWords(map[string]string{
		"白":  "白色",
		"玄关": "门厅",
	}))
	s := c.SynonymWords().String()
	assert.Equal(t, "白色门厅", s)
}

func TestCleaner_All(t *testing.T) {
	q := " 复古 未来主义的 ArT 🚩DECO 灵感设计客厅搭配复古未来的 北欧  👏 \n 现代       沙发🛋  "
	c := NewCleaner(q,
		WithStopWords("设计", "灵感"),
		WithCombineWords(
			[]string{"art deco"},
			map[string]string{
				"复古 未来": "复古未来",
			}),
		WithSynonymWords(map[string]string{
			"复古未来":         "复古未来主义",
			"setting-wall": "装饰墙",
		}),
	)
	assert.Equal(t, "复古未来主义主义的 art-deco 客厅搭配复古未来主义的 北欧 现代 沙发", c.Clean())
}

func BenchmarkCleaner_Clean(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		q := " 复古 未来主义的 ArT 🚩DECO 灵感设计客厅搭配复古未来的 北欧  👏 \n 现代       沙发🛋  "
		c := NewCleaner(q,
			WithStopWords("设计", "灵感"),
			WithCombineWords(
				[]string{"art deco"},
				map[string]string{
					"复古 未来": "复古未来",
				}),
			WithSynonymWords(map[string]string{
				"复古未来":         "复古未来主义",
				"setting-wall": "装饰墙",
			}),
		)
		c.Clean()
		// assert.Equal(b, "复古未来主义主义的 art-deco 客厅搭配复古未来主义的 北欧 现代 沙发", c.Clean())
	}
}
