package search

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

	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v2"
)

type FrontMatter struct {
	Title    string
	Subtitle string
	Summary  string
	Date     string
	Tags     []string
}

type Post struct {
	FrontMatter FrontMatter
	Content     string
	Path        string
	Uri         string
	Md5         string
	tokens      []string
}

var (
	HtmlReg, _  = regexp.Compile("<.{0,200}?>")
	PointReg, _ = regexp.Compile("[\n\t\r]")
)

func parsePost(dir, path string) (*Post, error) {
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

func ParsePost(nJob int, root string, sections ...string) chan *Post {
	files := make(map[string]struct{})

	contentDir := filepath.Join(root, "content")
	for _, section := range sections {
		dir := filepath.Join(contentDir, section)
		filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if filepath.Ext(path) == ".md" {
				files[path] = struct{}{}
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
				post, err := parsePost(contentDir, path)
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
