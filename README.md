# hugo_algolia_search

Due to algolia record size limit, we need to chunk our post to reduce index record size.


# How to use?
1. Build your hugo site, e.g. `~/blog`
2. In your hugo config, e.g. `~/blog/config.toml`, set like this:
    ```
    [params]
    algoliaAppId = "*****"
    algoliaApiKey = "**********"
    algoliaIndexName = "****"
    ```
3. Write some posts, e.g. `~/blog/content/history/a.md`, here `history` is a section.
4. Run following golang code
    ```go
    package main

    import "github.com/yinyajun/hugo_algolia_search"

    func main() {
    	search.Root = "~/blog"
    	search.BuildIndex("history")
    }
    ```
