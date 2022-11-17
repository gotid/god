package main

import (
	"fmt"
	"github.com/gotid/god/lib/stringx"
)

func main() {
	trie := stringx.NewTrie([]string{
		"AV演员",
		"AV女优",
		"苍井空",
		"色情",
	}, stringx.WithMask('?'))
	sentence, keywords, found := trie.Filter("日本AV演员兼电视、电影演员。苍井空AV女优是xx出道, 日本AV女优们最精彩的表演是AV演员色情表演")
	fmt.Println(sentence)
	fmt.Println(keywords)
	fmt.Println(found)
}
