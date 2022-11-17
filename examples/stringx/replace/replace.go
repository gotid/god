package main

import (
	"fmt"
	"strings"
)

func main() {
	txt := strings.NewReplacer("日本", "法国", "日本的首都", "东京", "东京", "日本的首都").
		Replace("日本的首都是东京")
	fmt.Println(txt)
}
