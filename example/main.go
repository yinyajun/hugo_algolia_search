package main

import "github.com/yinyajun/hugo_algolia_search"

func main() {
	print(123)
	search.Root = "/Users/yinyajun/Projects/github/yinyajun.github.io"
	search.BuildIndex("posts")
}

