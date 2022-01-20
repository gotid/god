package stringx

import (
	"git.zc0901.com/go/god/lib/lang"
)

const defaultMask = '*'

type (
	// Trie 定义了一个单词查找树接口。
	Trie interface {
		Filter(s string) (string, []string, bool)
		FilterWithFn(text string, fn func(string) string) (sentence string)
		FindKeywords(s string) []string
	}

	// TrieOption 是 Trie 的自定义函数。
	TrieOption func(trie *trieNode)

	trieNode struct {
		node
		mask rune
	}

	scope struct {
		start int
		stop  int
	}
)

// NewTrie 返回一个 Trie.
func NewTrie(words []string, opts ...TrieOption) Trie {
	n := new(trieNode)

	for _, opt := range opts {
		opt(n)
	}
	if n.mask == 0 {
		n.mask = defaultMask
	}
	for _, word := range words {
		n.add(word)
	}

	return n
}

func (n *trieNode) Filter(text string) (sentence string, keywords []string, found bool) {
	chars := []rune(text)
	if len(chars) == 0 {
		return text, nil, false
	}

	scopes := n.findKeywordScopes(chars)
	keywords = n.collectKeywords(chars, scopes)

	for _, match := range scopes {
		// we don't care about overlaps, not bringing a performance improvement
		n.replaceWithMask(chars, match.start, match.stop)
	}

	return string(chars), keywords, len(keywords) > 0
}

func (n *trieNode) FilterWithFn(text string, fn func(string) string) (sentence string) {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	scopes := n.findKeywordScopes(runes)
	replace := make(map[string]string)

	for _, score := range scopes {
		part := string(runes[score.start:score.stop])
		replace[part] = fn(part)
	}
	text = NewReplacer(replace).Replace(text)

	return text
}

func (n *trieNode) FindKeywords(text string) []string {
	chars := []rune(text)
	if len(chars) == 0 {
		return nil
	}

	scopes := n.findKeywordScopes(chars)
	return n.collectKeywords(chars, scopes)
}

func (n *trieNode) collectKeywords(chars []rune, scopes []scope) []string {
	set := make(map[string]lang.PlaceholderType)
	for _, v := range scopes {
		set[string(chars[v.start:v.stop])] = lang.Placeholder
	}

	var i int
	keywords := make([]string, len(set))
	for k := range set {
		keywords[i] = k
		i++
	}

	return keywords
}

func (n *trieNode) findKeywordScopes(chars []rune) []scope {
	var scopes []scope
	size := len(chars)
	start := -1

	for i := 0; i < size; i++ {
		child, ok := n.children[chars[i]]
		if !ok {
			continue
		}

		if start < 0 {
			start = i
		}
		if child.end {
			scopes = append(scopes, scope{
				start: start,
				stop:  i + 1,
			})
		}

		for j := i + 1; j < size; j++ {
			grandchild, ok := child.children[chars[j]]
			if !ok {
				break
			}

			child = grandchild
			if child.end {
				scopes = append(scopes, scope{
					start: start,
					stop:  j + 1,
				})
			}
		}

		start = -1
	}

	return scopes
}

func (n *trieNode) replaceWithMask(chars []rune, start, stop int) {
	for i := start; i < stop; i++ {
		chars[i] = n.mask
	}
}

// WithMask 自定义关键词掩码。
func WithMask(mask rune) TrieOption {
	return func(n *trieNode) {
		n.mask = mask
	}
}
