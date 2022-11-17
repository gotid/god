package stringx

import "github.com/gotid/god/lib/lang"

const defaultMask = '*'

type (
	// Trie 是一种快速查找元素的树实现。
	Trie interface {
		// Filter 使用掩码过滤 text 中的关键词。
		Filter(text string) (sentence string, keywords []string, found bool)
		// FindKeywords 查找 text 的中关键词。
		FindKeywords(text string) (keywords []string)
	}

	// TrieOption 自定义 Trie 的选项。
	TrieOption func(trie *trieNode)

	trieNode struct {
		node
		mask rune // 掩码字符
	}
)

// NewTrie 返回一个 Trie。
// keywords 为待过滤或查找的关键词。
func NewTrie(keywords []string, opts ...TrieOption) Trie {
	n := new(trieNode)

	for _, opt := range opts {
		opt(n)
	}
	if n.mask == 0 {
		n.mask = defaultMask
	}
	for _, word := range keywords {
		n.add(word)
	}

	n.build()

	return n
}

func (n *trieNode) Filter(text string) (sentence string, keywords []string, found bool) {
	chars := []rune(text)
	if len(chars) == 0 {
		return text, nil, false
	}

	scopes := n.find(chars)
	keywords = n.collectKeywords(chars, scopes)

	for _, match := range scopes {
		// 我们不关心重叠的关键词，因为不会带来性能的提升
		n.replaceWithMask(chars, match.start, match.stop)
	}

	return string(chars), keywords, len(keywords) > 0
}

func (n *trieNode) FindKeywords(text string) []string {
	chars := []rune(text)
	if len(chars) == 0 {
		return nil
	}

	scopes := n.find(chars)
	return n.collectKeywords(chars, scopes)
}

// 收集去重后的唯一关键词
func (n *trieNode) collectKeywords(chars []rune, scopes []scope) []string {
	set := make(map[string]lang.PlaceholderType)
	for _, s := range scopes {
		set[string(chars[s.start:s.stop])] = lang.Placeholder
	}

	var i int
	keywords := make([]string, len(set))
	for k := range set {
		keywords[i] = k
		i++
	}

	return keywords
}

// 使用掩码替换给定起止范围的字符。
func (n *trieNode) replaceWithMask(chars []rune, start int, stop int) {
	for i := start; i < stop; i++ {
		chars[i] = n.mask
	}
}

// WithMask 自定义替换关键词的掩码字符。
func WithMask(mask rune) TrieOption {
	return func(n *trieNode) {
		n.mask = mask
	}
}
