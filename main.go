package search

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

var (
	Root string
	NJob int = 4
)

type HugoConf struct {
	Params map[string]interface{}
}

func BuildIndex(sections ...string) {
	if Root == "" {
		panic("Root(hugo site) is not set")
	}

	// hugo conf parse
	var conf HugoConf

	if _, err := toml.DecodeFile(filepath.Join(Root, "config.toml"), &conf); err != nil {
		panic(err)
	}

	// walk hugo sections to parse posts
	posts := ParsePost(NJob, Root, sections...)

	for post := range posts {
		UpdateAlgoliaIndex(post, conf)
	}
}
