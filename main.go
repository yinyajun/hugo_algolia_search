package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

var (
	HtmlReg, _  = regexp.Compile("<.{0,200}?>")
	PointReg, _ = regexp.Compile("[\n\t\r]")

	conf HugoConf
)

type HugoConf struct {
	Params map[string]interface{}
}

func parseHugoConf(file string) {
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		panic(err)
	}
}

type FrontMatter struct {
	Title       string
	Subtitle    string
	Description string
	Summary     string
	Date        string
	Tags        []string
}

type Post struct {
	FrontMatter FrontMatter
	Content     string
	Path        string
	Uri         string
	Md5         string
	tokens      []string
}

func parsePost(path, dir string) (*Post, error) {
	b, _ := os.ReadFile(path)
	res := bytes.SplitN(b, []byte("---"), 3)

	// parse frontmatter
	var frontMatter FrontMatter
	err := yaml.Unmarshal(res[1], &frontMatter)
	if err != nil {
		return nil, err
	}

	// parse content
	content := blackfriday.Run(res[2])
	content = HtmlReg.ReplaceAll(content, []byte{'.'})
	content = PointReg.ReplaceAll(content, []byte{'.'})

	return &Post{
		FrontMatter: frontMatter,
		Content:     string(content),
		Path:        path,
		Uri:         strings.ReplaceAll(strings.ReplaceAll(path, dir, ""), ".md", ""),
		Md5:         fmt.Sprintf("%x", md5.Sum(content))}, nil
}

func ParsePost(nJob int, dirs ...string) chan *Post {
	files := make(map[string]string)
	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if filepath.Ext(path) == ".md" {
				files[path] = dir
			}
			return nil
		})
	}

	posts := make(chan *Post)

	go func() {
		limiter := make(chan struct{}, nJob)
		var wg sync.WaitGroup

		for path := range files {
			wg.Add(1)
			limiter <- struct{}{}
			go func(path string) {
				post, err := parsePost(path, files[path])
				if err != nil {
					log.Printf("parse %s failed: %s\n", path, err.Error())
				}
				posts <- post
				wg.Done()
				<-limiter
			}(path)
		}
		wg.Wait()
		close(limiter)
		close(posts)
	}()

	return posts
}

func main() {
	hugoConf := "/Users/yinyajun/Projects/github/yinyajun.github.io/config.toml"
	parseHugoConf(hugoConf)
	postDir := "/Users/yinyajun/Projects/github/yinyajun.github.io/content"
	posts := ParsePost(3, postDir)
	for post := range posts {
		UpdateAlgoliaIndex(post)
	}
}
