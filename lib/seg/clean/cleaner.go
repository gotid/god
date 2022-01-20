package clean

import (
	"strings"
	"unicode/utf8"

	"github.com/forPelevin/gomoji"

	"github.com/gotid/god/lib/stringx"

	emoji "github.com/tmdvs/Go-Emoji-Utils"
)

type (
	// Cleaner 表示一个查询词清理器
	Cleaner interface {
		String() string
		// Clean 全量清洗
		Clean() string
		// Synonym 返回同义词
		Synonym(string) string
		// Query 设置查询词
		Query(string) Cleaner
		// RemoveEmoji 去除表情
		RemoveEmoji() Cleaner
		// RemoveSpace 去除空格
		RemoveSpace() Cleaner
		// StopWords 停止词
		StopWords() Cleaner
		// CombineWords 合成词
		CombineWords() Cleaner
		// SynonymWords 同义词
		SynonymWords() Cleaner
	}

	// Option 是一个自定义选项函数
	Option func(cleaner *cleaner)

	cleaner struct {
		// 查询词
		query string
		// 删除词树
		stopWordTrie stringx.Trie
		// 合成词树
		combineWordTrie stringx.Trie
		// 合成词映射(用于自定义映射)
		combineMap map[string]string
		// 同义词树
		synonymWordTrie stringx.Trie
		// 同义词映射
		synonymMap map[string]string
	}
)

var _ emoji.Emoji

var _ Cleaner = (*cleaner)(nil)

// NewCleaner 返回一个新的查询词清理器。
func NewCleaner(query string, opts ...Option) Cleaner {
	c := new(cleaner)
	c.query = strings.ToLower(query)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Query 自定义停止词选项。
func (c *cleaner) Query(s string) Cleaner {
	c.query = strings.ToLower(s)
	return c
}

func (c *cleaner) Clean() string {
	return c.RemoveEmoji().
		RemoveSpace().
		StopWords().
		CombineWords().
		// SynonymWords().
		String()
}

func (c *cleaner) String() string {
	c.RemoveSpace()
	return c.query
}

func (c *cleaner) RemoveEmoji() Cleaner {
	// TODO 效果好但性能差
	// c.query = emoji.RemoveAll(c.query)

	// TODO 效果差但性能好，很多变种表情无法移除
	emojis := gomoji.FindAll(c.query)
	for _, e := range emojis {
		r, _ := utf8.DecodeRune([]byte(e.Character))
		c.query = strings.ReplaceAll(c.query, string([]rune{r}), "")
	}
	c.query = strings.TrimSpace(c.query)

	return c
}

func (c *cleaner) StopWords() Cleaner {
	if c.stopWordTrie == nil {
		return c
	}

	if output, _, ok := c.stopWordTrie.Filter(c.query); ok {
		c.query = output
	}

	return c
}

func (c *cleaner) CombineWords() Cleaner {
	if c.combineWordTrie == nil {
		return c
	}

	c.query = c.combineWordTrie.FilterWithFn(c.query, func(s string) string {
		// 有自定义合成词
		if c.combineMap != nil {
			if v, ok := c.combineMap[s]; ok {
				return v
			}
		}

		// 默认用-合成
		return strings.ReplaceAll(s, " ", "-")
	})

	return c
}

func (c *cleaner) SynonymWords() Cleaner {
	if c.synonymWordTrie == nil || c.synonymMap == nil {
		return c
	}

	c.query = c.synonymWordTrie.FilterWithFn(c.query, func(s string) string {
		if w, ok := c.synonymMap[s]; ok {
			return w
		}
		return s
	})

	return c
}

func (c *cleaner) Synonym(word string) string {
	if v, exists := c.synonymMap[word]; exists {
		return v
	}
	return word
}

func (c *cleaner) RemoveSpace() Cleaner {
	c.query = strings.Join(strings.Fields(c.query), " ")
	return c
}

// WithStopWords 自定义停止词选项。
func WithStopWords(words ...string) Option {
	return func(cleaner *cleaner) {
		cleaner.stopWordTrie = stringx.NewTrie(words, stringx.WithMask(' '))
	}
}

// WithCombineWords 自定义合成词选项。
func WithCombineWords(words []string, m ...map[string]string) Option {
	return func(cleaner *cleaner) {
		if len(m) > 0 {
			// 自定义合成词
			cleaner.combineMap = m[0]
			for k := range m[0] {
				if !stringx.Contains(words, k) {
					words = append(words, k)
				}
			}
		}
		cleaner.combineWordTrie = stringx.NewTrie(words)
	}
}

// WithSynonymWords 自定义同义词选项。
func WithSynonymWords(m map[string]string) Option {
	return func(cleaner *cleaner) {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		cleaner.synonymWordTrie = stringx.NewTrie(keys)
		cleaner.synonymMap = m
	}
}
