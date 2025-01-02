package main

import (
	"fmt"
	"github.com/yinyajun/hugo_algolia_search"
)

func main() {
	fmt.Println("======")
	search.Root = "/Users/yinyajun/Projects/github/yinyajun.github.io"
	search.BuildIndex("posts")
}
