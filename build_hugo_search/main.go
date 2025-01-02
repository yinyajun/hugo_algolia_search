package main

import (
	"flag"
	"fmt"
	"github.com/yinyajun/hugo_algolia_search"
)

func main() {
	fmt.Println("Begin to Build Hugo Algolia Search Index")
	root := flag.String("root_dir", ".", "Root directory of Hugo Blog")
	post := flag.String("post_dir", "posts", "Post directory of posts")

	flag.Parse()

	search.Root = *root
	search.BuildIndex(*post)
}