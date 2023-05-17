package search

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"unicode"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
)

var existed = make(map[string][]string)

func restore() string {
	if Root == "" {
		panic("Root(hugo site) is not set")
	}
	restoreFile := filepath.Join(Root, ".algolia_idx.ckpt")
	bytes, _ := os.ReadFile(restoreFile)
	json.Unmarshal(bytes, &existed)
	return restoreFile
}

type algoliaObject struct {
	ObjectID     string `json:"objectID"`
	Relpermalink string `json:"relpermalink"`
	Summary      string `json:"summary"`
	URL          string `json:"url"`
	Chunk        string `json:"chunk"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Content      string `json:"content"`
}

func chunk(content string) []string {
	var chunks []string
	var runes []rune
	var sz int

	for _, r := range content {
		if unicode.IsPunct(r) {
			continue
		}
		if unicode.IsDigit(r) {
			continue
		}
		runes = append(runes, r)
		sz += len(string(r))
	}
	num := math.Ceil(float64(sz) / 9400) // because of algolia record limit is 10k

	if num <= 1 {
		chunks = append(chunks, string(runes))
		return chunks
	}

	chunkSize := len(runes) / int(num)
	for i := 0; i <= int(num)-1; i++ {
		begin := i * chunkSize
		end := begin + chunkSize - 1
		if i == int(num)-1 {
			end = len(runes) - 1
		}
		chunks = append(chunks, string(runes[begin:end]))
	}
	return chunks
}

func formatObjects(post *Post) []algoliaObject {
	//chunks := chunk2(post)
	chunks := chunk(post.Content)

	var objects []algoliaObject
	for idx, ch := range chunks {
		obj := algoliaObject{
			ObjectID:     fmt.Sprintf("%s_%d", post.Uri, idx),
			Relpermalink: post.Uri,
			URL:          post.Uri,
			Chunk:        ch,
			Title:        post.FrontMatter.Title,
			Summary:      post.FrontMatter.Summary,
		}
		objects = append(objects, obj)
	}
	return objects
}

func UpdateAlgoliaIndex(post *Post, conf HugoConf) {
	f := restore()

	client := search.NewClient(conf.Params["algoliaAppId"].(string), conf.Params["algoliaApiKey"].(string))
	index := client.InitIndex(conf.Params["algoliaIndexName"].(string))

	objs := formatObjects(post)

	// delete old
	if len(existed[post.Uri]) > 0 {
		log.Printf("delete index of %s: %v", post.Uri, existed[post.Uri])
		index.DeleteObjects(existed[post.Uri])
	}
	// save new
	res, err := index.SaveObjects(objs)
	if err != nil {
		log.Println("save algolia index failed:", err)
		return
	}
	res.Wait()

	var ids []string
	for _, obj := range objs {
		ids = append(ids, obj.ObjectID)
	}
	log.Printf("save algolia index of %s: %s", post.Uri, ids)
	existed[post.Uri] = ids

	// save current
	bytes, _ := json.Marshal(existed)
	os.WriteFile(f, bytes, 0666)
}
